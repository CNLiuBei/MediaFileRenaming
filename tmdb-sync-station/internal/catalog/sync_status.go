package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *Store) SyncStatus(ctx context.Context) (SyncStatus, error) {
	status := SyncStatus{State: "idle", NextSync: "启动后自动同步", Description: "等待 TMDB 增量同步"}
	var finished, lastSuccess sql.NullTime
	var state, stage, message string
	err := s.db.QueryRowContext(ctx, `SELECT state,finished_at,stage,current,total,error,
		(SELECT max(finished_at) FROM sync_runs WHERE state='success')
		FROM sync_runs ORDER BY id DESC LIMIT 1`).Scan(&state, &finished, &stage, &status.Current, &status.Total, &message, &lastSuccess)
	if err == sql.ErrNoRows {
		return status, nil
	}
	if err != nil {
		return SyncStatus{}, fmt.Errorf("读取同步状态失败: %w", err)
	}
	status.State, status.Stage = state, stage
	if status.Total > 0 {
		status.Progress = int(status.Current * 100 / status.Total)
		if status.Progress > 100 {
			status.Progress = 100
		}
	}
	if lastSuccess.Valid {
		status.LastSync = lastSuccess.Time.Local().Format("2006-01-02 15:04")
	}
	if finished.Valid && state == "success" {
		status.NextSync = finished.Time.Add(24 * time.Hour).Local().Format("2006-01-02 15:04")
	}
	switch state {
	case "running":
		status.Description = stage
	case "failed":
		status.Description = message
	default:
		status.Description = "同步已完成"
	}
	return status, nil
}
