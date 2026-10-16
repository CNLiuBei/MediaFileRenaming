package syncer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"tmdb-sync-station/internal/catalog"
)

func (m *Manager) syncWindow(ctx context.Context) (time.Time, time.Time, bool, error) {
	end := dateOnly(time.Now().UTC())
	latestStart := end.AddDate(0, 0, -13)
	cursor, found, checkpoint, err := m.store.SyncCursor(ctx)
	if err != nil {
		return time.Time{}, time.Time{}, false, err
	}
	start, clipped := latestStart, !found
	if found {
		start = dateOnly(cursor)
		if checkpoint {
			start = start.AddDate(0, 0, 1)
		}
		clipped = start.Before(latestStart)
		if clipped {
			start = latestStart
		}
	}
	return start, end, clipped, nil
}

func (m *Manager) changedItems(ctx context.Context, runID int64, batch catalog.SyncBatch) (catalog.SyncBatch, error) {
	for !batch.ManifestComplete {
		kind, displayName, page := "", "", 0
		switch {
		case batch.MoviePage <= batch.MovieTotalPages:
			kind, displayName, page = "movie", "电影", batch.MoviePage
		case batch.TVPage <= batch.TVTotalPages:
			kind, displayName, page = "tv", "剧集", batch.TVPage
		default:
			return batch, nil
		}
		stage := fmt.Sprintf("读取%s变更清单，第 %d 页", displayName, page)
		if err := m.store.UpdateSyncRun(ctx, runID, 0, 0, stage); err != nil {
			return catalog.SyncBatch{}, err
		}
		query := url.Values{
			"start_date": {batch.WindowStart.Format(time.DateOnly)},
			"end_date":   {batch.WindowEnd.Format(time.DateOnly)},
			"page":       {strconv.Itoa(page)},
		}
		var result changedIDsPage
		if err := m.client.get(ctx, kind+"/changes", query, &result); err != nil {
			return catalog.SyncBatch{}, fmt.Errorf("读取%s变更清单失败: %w", displayName, err)
		}
		ids := make([]int64, 0, len(result.Results))
		for _, item := range result.Results {
			ids = append(ids, item.ID)
		}
		nextBatch, err := m.store.SaveSyncManifestPage(ctx, batch.ID, kind, page, result.TotalPages, len(result.Results) == 0, ids)
		if err != nil {
			return catalog.SyncBatch{}, fmt.Errorf("保存%s变更清单第 %d 页失败: %w", displayName, page, err)
		}
		batch = nextBatch
	}
	return batch, nil
}

func (m *Manager) syncItem(ctx context.Context, batchID int64, ref changeRef, runID, processed, total int64) error {
	query := url.Values{"language": {m.config.Language}, "append_to_response": {"credits,external_ids,alternative_titles"}}
	var details mediaDetails
	if err := m.client.get(ctx, fmt.Sprintf("%s/%d", ref.kind, ref.id), query, &details); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.NotFound {
			return m.store.DeleteSyncTitle(ctx, batchID, ref.kind, ref.id)
		}
		return fmt.Errorf("读取 TMDB %s %d 详情失败: %w", ref.kind, ref.id, err)
	}
	if details.ID != ref.id {
		return fmt.Errorf("TMDB 返回条目 ID 不匹配：请求 %d，得到 %d", ref.id, details.ID)
	}
	title, err := convertDetails(details, ref.kind)
	if err != nil {
		return err
	}
	seasons := make([]catalog.SyncSeason, 0, len(details.Seasons))
	if ref.kind == "tv" {
		if details.NumberOfSeasons > 0 && len(details.Seasons) == 0 {
			return fmt.Errorf("TMDB 未返回剧集 %d 的分季列表", ref.id)
		}
		for index, summary := range details.Seasons {
			if summary.Number < 0 {
				continue
			}
			stage := fmt.Sprintf("同步剧集 %d：读取第 %d / %d 季", ref.id, index+1, len(details.Seasons))
			if err := m.store.UpdateSyncRun(ctx, runID, processed, total, stage); err != nil {
				return err
			}
			var season seasonDetails
			if err := m.client.get(ctx, fmt.Sprintf("tv/%d/season/%d", ref.id, summary.Number), url.Values{"language": {m.config.Language}}, &season); err != nil {
				return fmt.Errorf("读取 TMDB 剧集 %d 第 %d 季失败: %w", ref.id, summary.Number, err)
			}
			entry := catalog.SyncSeason{Number: summary.Number, Name: season.Name, EpisodeCount: summary.EpisodeCount, Episodes: make([]catalog.SyncEpisode, 0, len(season.Episodes))}
			if entry.Name == "" {
				entry.Name = summary.Name
			}
			if len(season.Episodes) > 0 {
				entry.EpisodeCount = len(season.Episodes)
			}
			for _, episode := range season.Episodes {
				entry.Episodes = append(entry.Episodes, catalog.SyncEpisode{Number: episode.Number, Title: episode.Name})
			}
			seasons = append(seasons, entry)
		}
	}
	if _, err := m.store.ApplySyncTitle(ctx, batchID, title, seasons); err != nil {
		return fmt.Errorf("写入 PostgreSQL 条目 %s %d 失败: %w", ref.kind, ref.id, err)
	}
	return nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
