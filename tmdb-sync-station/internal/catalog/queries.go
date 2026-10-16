package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

func (s *Store) Dashboard(ctx context.Context) (Dashboard, error) {
	var result Dashboard
	err := s.db.QueryRowContext(ctx, `SELECT
	(SELECT count(*) FROM titles),
	(SELECT count(*) FROM titles WHERE kind='movie'), (SELECT count(*) FROM titles WHERE kind='tv'),
	(SELECT count(*) FROM episodes), (SELECT count(*) FROM aliases),
	pg_size_pretty(pg_database_size(current_database()))`).Scan(
		&result.Catalog.Titles, &result.Catalog.Movies, &result.Catalog.Series,
		&result.Catalog.Episodes, &result.Catalog.Aliases, &result.Catalog.Size)
	if err != nil {
		return Dashboard{}, fmt.Errorf("读取片库统计失败: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT
	COALESCE((SELECT value FROM catalog_meta WHERE key='source'),''),
	COALESCE((SELECT to_char(to_timestamp(value::double precision), 'YYYY-MM-DD HH24:MI:SS') FROM catalog_meta WHERE key='updated_at'),''),
	COALESCE((SELECT value FROM catalog_meta WHERE key='imported_at'),'')`).Scan(
		&result.Catalog.Source, &result.Catalog.SourceUpdatedAt, &result.Catalog.ImportedAt); err != nil {
		return Dashboard{}, fmt.Errorf("读取片库来源信息失败: %w", err)
	}
	result.Sync, err = s.SyncStatus(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	if result.Recent, err = s.recentChanges(ctx); err != nil {
		return Dashboard{}, err
	}
	if result.Daily, err = s.dailyChanges(ctx); err != nil {
		return Dashboard{}, err
	}
	if result.Runs, err = s.SyncRuns(ctx, 20); err != nil {
		return Dashboard{}, err
	}
	return result, nil
}

func (s *Store) recentChanges(ctx context.Context) ([]Change, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT c.tmdb_id, COALESCE(t.title, '未知条目'),
	CASE c.media_type WHEN 'movie' THEN '电影' ELSE '电视剧' END, c.action, c.changed_at
	FROM sync_changes c LEFT JOIN titles t ON c.media_type=t.kind AND t.tmdb_id=c.tmdb_id
	ORDER BY c.changed_at DESC LIMIT 8`)
	if err != nil {
		return nil, fmt.Errorf("读取最近变更失败: %w", err)
	}
	defer rows.Close()
	changes := make([]Change, 0)
	for rows.Next() {
		var item Change
		var updated time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.Kind, &item.Action, &updated); err != nil {
			return nil, err
		}
		item.UpdatedAt = updated.Local().Format("2006-01-02 15:04")
		changes = append(changes, item)
	}
	return changes, rows.Err()
}

func (s *Store) dailyChanges(ctx context.Context) ([]DailyCount, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT date_trunc('day', changed_at)::date, count(*) FROM sync_changes
	WHERE changed_at >= current_date - interval '6 days' GROUP BY 1 ORDER BY 1`)
	if err != nil {
		return nil, fmt.Errorf("读取每日变更量失败: %w", err)
	}
	defer rows.Close()
	counts := make([]DailyCount, 0, 7)
	for rows.Next() {
		var day time.Time
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		counts = append(counts, DailyCount{Day: day.Format("01-02"), Count: count})
	}
	return counts, rows.Err()
}

func (s *Store) SyncRuns(ctx context.Context, limit int) ([]SyncRun, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, started_at, finished_at, state, processed, current, total, stage, error FROM sync_runs ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("读取同步历史失败: %w", err)
	}
	defer rows.Close()
	runs := make([]SyncRun, 0)
	for rows.Next() {
		var run SyncRun
		var finished sql.NullTime
		var started time.Time
		if err := rows.Scan(&run.ID, &started, &finished, &run.State, &run.Processed, &run.Current, &run.Total, &run.Stage, &run.Error); err != nil {
			return nil, err
		}
		run.StartedAt = started.Local().Format("2006-01-02 15:04:05")
		if finished.Valid {
			run.FinishedAt = finished.Time.Local().Format("2006-01-02 15:04:05")
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *Store) Catalog(ctx context.Context, kind, search string, limit, offset int) ([]CatalogItem, bool, error) {
	if limit < 1 || limit > 100 {
		limit = 25
	}
	kind = map[string]string{"全部": "", "电影": "movie", "电视剧": "tv"}[kind]
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if kind != "" {
		args = append(args, kind)
		conditions = append(conditions, fmt.Sprintf("t.kind=$%d", len(args)))
	}
	prefixSearch := search != "" && utf8.RuneCountInString(search) < 3
	prefixPlaceholder := ""
	if search != "" {
		if prefixSearch {
			args = append(args, escapeLikePattern(strings.ToLower(search))+"%")
			prefixPlaceholder = fmt.Sprintf("$%d", len(args))
		} else {
			args = append(args, "%"+search+"%")
			conditions = append(conditions, fmt.Sprintf("(t.title || ' ' || t.original_title || ' ' || t.english_title) ILIKE $%d", len(args)))
		}
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	selectColumns := `SELECT t.tmdb_id,t.title,t.original_title,t.overview,COALESCE(t.year::text,''),t.poster_path,t.popularity,t.updated_at,
	CASE t.kind WHEN 'movie' THEN '电影' ELSE '电视剧' END`
	query := selectColumns + ` FROM titles t`
	if prefixSearch {
		prefixCondition := `lower(title) LIKE ` + prefixPlaceholder + ` ESCAPE E'\\'`
		matchedTitles := `SELECT tmdb_id,kind FROM titles WHERE ` + prefixCondition +
			` UNION SELECT tmdb_id,kind FROM titles WHERE ` + prefixCondition +
			` UNION SELECT tmdb_id,kind FROM titles WHERE ` + prefixCondition
		query = `WITH matched AS MATERIALIZED (` + matchedTitles + `) ` + selectColumns +
			` FROM matched m JOIN titles t ON t.tmdb_id=m.tmdb_id AND t.kind=m.kind`
	}
	query += where
	args = append(args, limit+1, offset)
	query += fmt.Sprintf(" ORDER BY t.popularity DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("查询片库失败: %w", err)
	}
	defer rows.Close()
	items := make([]CatalogItem, 0)
	for rows.Next() {
		var item CatalogItem
		var updated time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.OriginalTitle, &item.Overview, &item.ReleaseDate, &item.PosterPath, &item.Popularity, &updated, &item.Kind); err != nil {
			return nil, false, err
		}
		item.UpdatedAt = updated.Local().Format("2006-01-02 15:04")
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	return items, hasMore, nil
}

func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
