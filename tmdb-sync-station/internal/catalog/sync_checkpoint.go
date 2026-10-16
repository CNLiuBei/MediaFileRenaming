package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type SyncBatch struct {
	ID               int64
	WindowStart      time.Time
	WindowEnd        time.Time
	Clipped          bool
	MoviePage        int
	MovieTotalPages  int
	TVPage           int
	TVTotalPages     int
	ManifestComplete bool
}

type SyncBatchItem struct {
	MediaType string
	TMDBID    int64
}

const syncBatchColumns = `id,window_start,window_end,clipped,movie_page,movie_total_pages,
	tv_page,tv_total_pages,manifest_complete`

func (s *Store) GetOrCreateSyncBatch(ctx context.Context, start, end time.Time, clipped bool) (SyncBatch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SyncBatch{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `LOCK TABLE sync_batches IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return SyncBatch{}, err
	}
	batch, err := scanSyncBatch(tx.QueryRowContext(ctx, `SELECT `+syncBatchColumns+` FROM sync_batches WHERE completed=false ORDER BY id LIMIT 1`))
	if err == sql.ErrNoRows {
		batch, err = scanSyncBatch(tx.QueryRowContext(ctx, `INSERT INTO sync_batches(window_start,window_end,clipped)
			VALUES($1,$2,$3) RETURNING `+syncBatchColumns, start, end, clipped))
	}
	if err != nil {
		return SyncBatch{}, fmt.Errorf("读取或创建同步检查点失败: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return SyncBatch{}, err
	}
	return batch, nil
}

func (s *Store) SaveSyncManifestPage(ctx context.Context, batchID int64, mediaType string, page, totalPages int, empty bool, ids []int64) (SyncBatch, error) {
	if mediaType != "movie" && mediaType != "tv" {
		return SyncBatch{}, fmt.Errorf("不支持的 TMDB 类型 %q", mediaType)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SyncBatch{}, err
	}
	defer tx.Rollback()

	pageColumn := "movie_page"
	if mediaType == "tv" {
		pageColumn = "tv_page"
	}
	var expectedPage int
	if err = tx.QueryRowContext(ctx, `SELECT `+pageColumn+` FROM sync_batches WHERE id=$1 AND completed=false FOR UPDATE`, batchID).Scan(&expectedPage); err != nil {
		return SyncBatch{}, fmt.Errorf("读取变更清单页检查点失败: %w", err)
	}
	if expectedPage > page {
		_ = tx.Rollback()
		return s.SyncBatch(ctx, batchID)
	}
	if expectedPage != page {
		return SyncBatch{}, fmt.Errorf("变更清单检查点不连续：期望第 %d 页，收到第 %d 页", expectedPage, page)
	}
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO sync_batch_items(batch_id,media_type,tmdb_id)
			VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, batchID, mediaType, id); err != nil {
			return SyncBatch{}, fmt.Errorf("保存变更影视 ID 失败: %w", err)
		}
	}
	if empty {
		totalPages = page - 1
	} else if totalPages < 1 {
		totalPages = 1
	}
	if mediaType == "movie" {
		if empty {
			_, err = tx.ExecContext(ctx, `UPDATE sync_batches SET movie_page=$2,movie_total_pages=$3 WHERE id=$1`, batchID, page+1, totalPages)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE sync_batches SET movie_page=$2,movie_total_pages=GREATEST(movie_total_pages,$3) WHERE id=$1`, batchID, page+1, totalPages)
		}
	} else {
		if empty {
			_, err = tx.ExecContext(ctx, `UPDATE sync_batches SET tv_page=$2,tv_total_pages=$3 WHERE id=$1`, batchID, page+1, totalPages)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE sync_batches SET tv_page=$2,tv_total_pages=GREATEST(tv_total_pages,$3) WHERE id=$1`, batchID, page+1, totalPages)
		}
	}
	if err != nil {
		return SyncBatch{}, err
	}
	batch, err := scanSyncBatch(tx.QueryRowContext(ctx, `SELECT `+syncBatchColumns+` FROM sync_batches WHERE id=$1`, batchID))
	if err != nil {
		return SyncBatch{}, err
	}
	if err = tx.Commit(); err != nil {
		return SyncBatch{}, err
	}
	return batch, nil
}

func (s *Store) SyncBatch(ctx context.Context, batchID int64) (SyncBatch, error) {
	batch, err := scanSyncBatch(s.db.QueryRowContext(ctx, `SELECT `+syncBatchColumns+` FROM sync_batches WHERE id=$1`, batchID))
	if err != nil {
		return SyncBatch{}, fmt.Errorf("读取同步检查点失败: %w", err)
	}
	return batch, nil
}

func (s *Store) CompleteSyncManifest(ctx context.Context, batchID int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_batches SET manifest_complete=true WHERE id=$1 AND completed=false`, batchID)
	return err
}

func (s *Store) CompleteSyncBatch(ctx context.Context, batchID int64, cursor time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO catalog_meta(key,value) VALUES('sync_cursor',$1)
		ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value`, cursor.Format(time.DateOnly)); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE sync_batches SET completed=true,completed_at=now()
		WHERE id=$1 AND manifest_complete=true AND NOT EXISTS
		(SELECT 1 FROM sync_batch_items WHERE batch_id=$1 AND completed=false)`, batchID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("同步检查点仍有未完成的变更项，不能提交水位")
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM sync_batch_items WHERE batch_id=$1`, batchID); err != nil {
		return err
	}
	return tx.Commit()
}

func scanSyncBatch(row *sql.Row) (SyncBatch, error) {
	var batch SyncBatch
	err := row.Scan(&batch.ID, &batch.WindowStart, &batch.WindowEnd, &batch.Clipped,
		&batch.MoviePage, &batch.MovieTotalPages, &batch.TVPage, &batch.TVTotalPages, &batch.ManifestComplete)
	return batch, err
}
