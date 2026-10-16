package catalog

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) SyncBatchItemCounts(ctx context.Context, batchID int64) (completed, total int64, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE completed),count(*) FROM sync_batch_items WHERE batch_id=$1`, batchID).Scan(&completed, &total)
	return completed, total, err
}

func (s *Store) PendingSyncBatchItems(ctx context.Context, batchID int64) ([]SyncBatchItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT media_type,tmdb_id FROM sync_batch_items
		WHERE batch_id=$1 AND completed=false ORDER BY media_type,tmdb_id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SyncBatchItem, 0)
	for rows.Next() {
		var item SyncBatchItem
		if err = rows.Scan(&item.MediaType, &item.TMDBID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func markSyncBatchItemComplete(ctx context.Context, tx *sql.Tx, batchID int64, mediaType string, tmdbID int64) error {
	result, err := tx.ExecContext(ctx, `UPDATE sync_batch_items SET completed=true
		WHERE batch_id=$1 AND media_type=$2 AND tmdb_id=$3`, batchID, mediaType, tmdbID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows > 0 {
		return err
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sync_batch_items WHERE batch_id=$1 AND media_type=$2 AND tmdb_id=$3)`, batchID, mediaType, tmdbID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("同步队列项不存在：%s %d", mediaType, tmdbID)
	}
	return nil
}
