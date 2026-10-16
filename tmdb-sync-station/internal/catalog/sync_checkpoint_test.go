package catalog

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSyncBatchCheckpointResume(t *testing.T) {
	dsn := os.Getenv("TMDB_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("需要独立的 TMDB_SYNC_TEST_DATABASE_URL PostgreSQL 测试库")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	store, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	var databaseName string
	if err = store.db.QueryRowContext(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("测试只能连接数据库名以 _test 结尾的隔离库，当前为 %q", databaseName)
	}
	var activeBatches int
	if err = store.db.QueryRowContext(ctx, `SELECT count(*) FROM sync_batches WHERE completed=false`).Scan(&activeBatches); err != nil {
		t.Fatal(err)
	}
	if activeBatches != 0 {
		t.Fatalf("测试库已有 %d 个未完成批次，请先清理隔离测试库", activeBatches)
	}

	var oldCursor sql.NullString
	err = store.db.QueryRowContext(ctx, `SELECT value FROM catalog_meta WHERE key='sync_cursor'`).Scan(&oldCursor)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	baseID := time.Now().UnixNano()
	start := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	batch, err := store.GetOrCreateSyncBatch(ctx, start, end, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM sync_changes WHERE tmdb_id BETWEEN $1 AND $2`, baseID, baseID+2)
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM titles WHERE tmdb_id BETWEEN $1 AND $2`, baseID, baseID+2)
		_, _ = store.db.ExecContext(context.Background(), `DELETE FROM sync_batches WHERE id=$1`, batch.ID)
		if oldCursor.Valid {
			_, _ = store.db.ExecContext(context.Background(), `UPDATE catalog_meta SET value=$1 WHERE key='sync_cursor'`, oldCursor.String)
		} else {
			_, _ = store.db.ExecContext(context.Background(), `DELETE FROM catalog_meta WHERE key='sync_cursor'`)
		}
		_ = store.Close()
	})

	batch, err = store.SaveSyncManifestPage(ctx, batch.ID, "movie", 1, 2, false, []int64{baseID, baseID + 1})
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := store.GetOrCreateSyncBatch(ctx, start, end, false)
	if err != nil {
		t.Fatal(err)
	}
	if resumed.ID != batch.ID || resumed.MoviePage != 2 || resumed.MovieTotalPages != 2 {
		t.Fatalf("恢复到错误的电影清单检查点：%+v", resumed)
	}
	batch, err = store.SaveSyncManifestPage(ctx, batch.ID, "movie", 2, 2, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	batch, err = store.SaveSyncManifestPage(ctx, batch.ID, "tv", 1, 1, false, []int64{baseID + 2})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.CompleteSyncManifest(ctx, batch.ID); err != nil {
		t.Fatal(err)
	}
	completed, total, err := store.SyncBatchItemCounts(ctx, batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed != 0 || total != 3 {
		t.Fatalf("清单项目数错误：已完成 %d，总计 %d", completed, total)
	}

	if _, err = store.ApplySyncTitle(ctx, batch.ID, SyncTitle{ID: baseID, Kind: "movie", Title: "checkpoint test"}, nil); err != nil {
		t.Fatal(err)
	}
	if err = store.DeleteSyncTitle(ctx, batch.ID, "movie", baseID+1); err != nil {
		t.Fatal(err)
	}
	if _, err = store.ApplySyncTitle(ctx, batch.ID, SyncTitle{ID: baseID + 2, Kind: "tv", Title: "checkpoint test"}, nil); err != nil {
		t.Fatal(err)
	}
	completed, total, err = store.SyncBatchItemCounts(ctx, batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed != 3 || total != 3 {
		t.Fatalf("条目完成状态错误：已完成 %d，总计 %d", completed, total)
	}
	if err = store.CompleteSyncBatch(ctx, batch.ID, end.AddDate(0, 0, -1)); err != nil {
		t.Fatal(err)
	}
	cursor, found, checkpoint, err := store.SyncCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !found || !checkpoint || !cursor.Equal(end.AddDate(0, 0, -1)) {
		t.Fatalf("同步水位没有正确提交：cursor=%s found=%t checkpoint=%t", cursor, found, checkpoint)
	}
}
