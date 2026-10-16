package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var ErrSyncAlreadyRunning = errors.New("已有同步任务正在运行")

func (s *Store) BeginSyncRun(ctx context.Context) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `INSERT INTO sync_runs(state, stage) VALUES ('running', '准备同步') RETURNING id`).Scan(&id)
	return id, err
}

func (s *Store) UpdateSyncRun(ctx context.Context, id, current, total int64, stage string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_runs SET current=$2,total=$3,processed=$2,stage=$4 WHERE id=$1 AND state='running'`, id, current, total, stage)
	return err
}

func (s *Store) FinishSyncRun(ctx context.Context, id, current, total int64, state, message, stage string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_runs SET state=$2,processed=$3,current=$3,total=$4,error=$5,stage=$6,finished_at=now() WHERE id=$1`, id, state, current, total, message, stage)
	return err
}

func (s *Store) MarkInterruptedSyncRuns(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_runs SET state='failed', finished_at=now(), error='服务在同步期间重启', stage='同步中断' WHERE state='running'`)
	return err
}

func (s *Store) SyncCursor(ctx context.Context) (time.Time, bool, bool, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM catalog_meta WHERE key='sync_cursor'`).Scan(&value)
	if err == nil {
		cursor, parseErr := time.Parse(time.DateOnly, value)
		return cursor, parseErr == nil, true, parseErr
	}
	if err != sql.ErrNoRows {
		return time.Time{}, false, false, fmt.Errorf("读取同步水位失败: %w", err)
	}
	err = s.db.QueryRowContext(ctx, `SELECT value FROM catalog_meta WHERE key='updated_at'`).Scan(&value)
	if err == sql.ErrNoRows {
		return time.Time{}, false, false, nil
	}
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("读取离线快照时间失败: %w", err)
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("解析离线快照时间失败: %w", err)
	}
	return time.Unix(int64(seconds), 0).UTC(), true, false, nil
}

func (s *Store) SaveSyncCursor(ctx context.Context, cursor time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO catalog_meta(key,value) VALUES ('sync_cursor',$1) ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value`, cursor.Format(time.DateOnly))
	return err
}
