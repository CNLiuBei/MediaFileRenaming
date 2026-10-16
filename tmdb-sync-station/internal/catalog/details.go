package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
)

func (s *Store) CatalogDetail(ctx context.Context, kind string, id int64) (CatalogDetail, error) {
	var item CatalogDetail
	var countries, castMembers, networks, companies []byte
	err := s.db.QueryRowContext(ctx, `SELECT t.tmdb_id,
		CASE t.kind WHEN 'movie' THEN '电影' ELSE '电视剧' END,
		t.title, t.original_title, t.overview, t.tagline, COALESCE(t.year, 0),
		t.original_language, t.origin_countries, t.poster_path, t.backdrop_path,
		t.popularity, t.vote_average, t.vote_count, COALESCE(t.runtime, 0), t.director,
		t.cast_members, t.networks, t.companies, t.status, t.homepage, t.imdb_id,
		t.season_count, t.episode_count, t.in_production, t.collection_name
		FROM titles t WHERE t.kind=$1 AND t.tmdb_id=$2`, kind, id).Scan(
		&item.ID, &item.Kind, &item.Title, &item.OriginalTitle, &item.Overview, &item.Tagline, &item.Year,
		&item.OriginalLanguage, &countries, &item.PosterPath, &item.BackdropPath,
		&item.Popularity, &item.VoteAverage, &item.VoteCount, &item.Runtime, &item.Director,
		&castMembers, &networks, &companies, &item.Status, &item.Homepage, &item.IMDbID,
		&item.SeasonCount, &item.EpisodeCount, &item.InProduction, &item.CollectionName)
	if err != nil {
		return CatalogDetail{}, err
	}
	item.OriginCountries = jsonOrEmptyArray(countries)
	item.CastMembers = jsonOrEmptyArray(castMembers)
	item.Networks = jsonOrEmptyArray(networks)
	item.Companies = jsonOrEmptyArray(companies)
	item.Seasons = make([]SeasonDetail, 0)
	if kind == "tv" {
		item.Seasons, err = s.catalogSeasons(ctx, id)
		if err != nil {
			return CatalogDetail{}, err
		}
	}
	return item, nil
}

func (s *Store) catalogSeasons(ctx context.Context, showID int64) ([]SeasonDetail, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT s.season_number,s.name,s.episode_count,e.episode,e.title
		FROM seasons s LEFT JOIN episodes e ON e.show_id=s.show_id AND e.season=s.season_number
		WHERE s.show_id=$1 ORDER BY s.season_number,e.episode`, showID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seasons := make([]SeasonDetail, 0)
	for rows.Next() {
		var number, episodeCount int
		var name string
		var episodeNumber sql.NullInt64
		var episodeTitle sql.NullString
		if err := rows.Scan(&number, &name, &episodeCount, &episodeNumber, &episodeTitle); err != nil {
			return nil, err
		}
		if len(seasons) == 0 || seasons[len(seasons)-1].SeasonNumber != number {
			seasons = append(seasons, SeasonDetail{SeasonNumber: number, Name: name, EpisodeCount: episodeCount, Episodes: make([]EpisodeDetail, 0)})
		}
		if episodeNumber.Valid {
			season := &seasons[len(seasons)-1]
			season.Episodes = append(season.Episodes, EpisodeDetail{EpisodeNumber: int(episodeNumber.Int64), Title: episodeTitle.String})
		}
	}
	return seasons, rows.Err()
}

func jsonOrEmptyArray(value []byte) json.RawMessage {
	if len(value) == 0 || string(value) == "null" {
		return json.RawMessage("[]")
	}
	if !json.Valid(value) {
		return json.RawMessage("[]")
	}
	return json.RawMessage(value)
}
