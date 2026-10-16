package catalog

import (
	"context"
	"encoding/json"
	"strings"
)

type SyncTitle struct {
	ID                                                                                                           int64
	Kind, Title, OriginalTitle, EnglishTitle, Overview, OriginalLanguage                                         string
	Year, Runtime, VoteCount, SeasonCount, EpisodeCount                                                          int
	Popularity, VoteAverage                                                                                      float64
	GenreIDs, OriginCountries, CastMembers, Networks, Companies                                                  json.RawMessage
	Director, PosterPath, BackdropPath, IMDbID, Tagline, Status, Homepage, LastAirDate, ShowType, CollectionName string
	InProduction                                                                                                 bool
	CollectionID                                                                                                 *int64
	Aliases                                                                                                      []string
}

type SyncSeason struct {
	Number       int
	Name         string
	EpisodeCount int
	Episodes     []SyncEpisode
}

type SyncEpisode struct {
	Number int
	Title  string
}

func (s *Store) ApplySyncTitle(ctx context.Context, batchID int64, title SyncTitle, seasons []SyncSeason) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM titles WHERE tmdb_id=$1 AND kind=$2)`, title.ID, title.Kind).Scan(&exists); err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO titles (
		tmdb_id,kind,title,original_title,english_title,year,overview,popularity,vote_average,original_language,
		genre_ids,origin_countries,director,cast_members,networks,companies,poster_path,backdrop_path,runtime,vote_count,
		imdb_id,tagline,status,homepage,last_air_date,season_count,episode_count,in_production,show_type,collection_id,collection_name,detail_version,updated_at
	) VALUES ($1,$2,$3,$4,$5,NULLIF($6,0),$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,NULLIF($19,0),$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,1,now())
	ON CONFLICT(tmdb_id,kind) DO UPDATE SET title=EXCLUDED.title,original_title=EXCLUDED.original_title,english_title=EXCLUDED.english_title,
	 year=EXCLUDED.year,overview=EXCLUDED.overview,popularity=EXCLUDED.popularity,vote_average=EXCLUDED.vote_average,
	 original_language=EXCLUDED.original_language,genre_ids=EXCLUDED.genre_ids,origin_countries=EXCLUDED.origin_countries,
	 director=EXCLUDED.director,cast_members=EXCLUDED.cast_members,networks=EXCLUDED.networks,companies=EXCLUDED.companies,
	 poster_path=EXCLUDED.poster_path,backdrop_path=EXCLUDED.backdrop_path,runtime=EXCLUDED.runtime,vote_count=EXCLUDED.vote_count,
	 imdb_id=EXCLUDED.imdb_id,tagline=EXCLUDED.tagline,status=EXCLUDED.status,homepage=EXCLUDED.homepage,
	 last_air_date=EXCLUDED.last_air_date,season_count=EXCLUDED.season_count,episode_count=EXCLUDED.episode_count,
	 in_production=EXCLUDED.in_production,show_type=EXCLUDED.show_type,collection_id=EXCLUDED.collection_id,
	 collection_name=EXCLUDED.collection_name,detail_version=1,updated_at=now()`,
		title.ID, title.Kind, title.Title, title.OriginalTitle, title.EnglishTitle, title.Year, title.Overview,
		title.Popularity, title.VoteAverage, title.OriginalLanguage, jsonValue(title.GenreIDs), jsonValue(title.OriginCountries),
		title.Director, jsonValue(title.CastMembers), jsonValue(title.Networks), jsonValue(title.Companies), title.PosterPath,
		title.BackdropPath, title.Runtime, title.VoteCount, title.IMDbID, title.Tagline, title.Status, title.Homepage,
		title.LastAirDate, title.SeasonCount, title.EpisodeCount, title.InProduction, title.ShowType, title.CollectionID, title.CollectionName)
	if err != nil {
		return "", err
	}
	for _, alias := range title.Aliases {
		name := strings.TrimSpace(alias)
		key := strings.ToLower(strings.Join(strings.Fields(name), " "))
		if key == "" {
			continue
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO aliases(tmdb_id,kind,name,name_key) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, title.ID, title.Kind, name, key); err != nil {
			return "", err
		}
	}
	if title.Kind == "tv" {
		if _, err = tx.ExecContext(ctx, `DELETE FROM episodes WHERE show_id=$1`, title.ID); err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM seasons WHERE show_id=$1`, title.ID); err != nil {
			return "", err
		}
		for _, season := range seasons {
			if _, err = tx.ExecContext(ctx, `INSERT INTO seasons(show_id,season_number,name,episode_count) VALUES($1,$2,$3,$4)`, title.ID, season.Number, season.Name, season.EpisodeCount); err != nil {
				return "", err
			}
			for _, episode := range season.Episodes {
				if _, err = tx.ExecContext(ctx, `INSERT INTO episodes(show_id,season,episode,title) VALUES($1,$2,$3,$4)`, title.ID, season.Number, episode.Number, episode.Title); err != nil {
					return "", err
				}
			}
		}
	}
	action := "更新"
	if !exists {
		action = "新增"
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO sync_changes(media_type,tmdb_id,action) VALUES($1,$2,$3)`, title.Kind, title.ID, action); err != nil {
		return "", err
	}
	if err = markSyncBatchItemComplete(ctx, tx, batchID, title.Kind, title.ID); err != nil {
		return "", err
	}
	return action, tx.Commit()
}

func (s *Store) DeleteSyncTitle(ctx context.Context, batchID int64, kind string, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM titles WHERE kind=$1 AND tmdb_id=$2`, kind, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		if kind == "tv" {
			if _, err = tx.ExecContext(ctx, `DELETE FROM episodes WHERE show_id=$1`, id); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `DELETE FROM seasons WHERE show_id=$1`, id); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO sync_changes(media_type,tmdb_id,action) VALUES($1,$2,'删除')`, kind, id); err != nil {
			return err
		}
	}
	if err = markSyncBatchItemComplete(ctx, tx, batchID, kind, id); err != nil {
		return err
	}
	return tx.Commit()
}

func jsonValue(value json.RawMessage) []byte {
	if len(value) == 0 || !json.Valid(value) {
		return []byte("[]")
	}
	return value
}
