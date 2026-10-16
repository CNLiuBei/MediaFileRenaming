package catalog

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Store struct{ db *sql.DB }

func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 PostgreSQL 连接失败: %w", err)
	}
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	pingCtx, cancelPing := context.WithTimeout(ctx, 10*time.Second)
	defer cancelPing()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接 PostgreSQL 失败: %w", err)
	}
	store := &Store{db: db}
	migrationCtx, cancelMigration := context.WithTimeout(ctx, 30*time.Minute)
	defer cancelMigration()
	if err := store.Migrate(migrationCtx); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Health(ctx context.Context) (map[string]string, error) {
	var database, user, version string
	if err := s.db.QueryRowContext(ctx, `SELECT current_database(), current_user, version()`).Scan(&database, &user, &version); err != nil {
		return nil, fmt.Errorf("检查 PostgreSQL 状态失败: %w", err)
	}
	return map[string]string{"state": "connected", "database": database, "user": user, "version": version}, nil
}

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("初始化迁移记录失败: %w", err)
	}
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("读取迁移目录失败: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		var applied bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, entry.Name()).Scan(&applied); err != nil {
			return fmt.Errorf("检查迁移 %s 失败: %w", entry.Name(), err)
		}
		if applied {
			continue
		}
		script, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("读取迁移 %s 失败: %w", entry.Name(), err)
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("开启迁移 %s 事务失败: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, string(script)); err != nil {
			tx.Rollback()
			return fmt.Errorf("执行迁移 %s 失败: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, entry.Name()); err != nil {
			tx.Rollback()
			return fmt.Errorf("记录迁移 %s 失败: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交迁移 %s 失败: %w", entry.Name(), err)
		}
	}
	return nil
}
