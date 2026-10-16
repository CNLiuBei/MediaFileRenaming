CREATE TABLE IF NOT EXISTS movies (
    tmdb_id BIGINT PRIMARY KEY, title TEXT NOT NULL, original_title TEXT NOT NULL DEFAULT '',
    overview TEXT NOT NULL DEFAULT '', release_date DATE, poster_path TEXT NOT NULL DEFAULT '',
    popularity DOUBLE PRECISION NOT NULL DEFAULT 0, raw_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS tv_series (
    tmdb_id BIGINT PRIMARY KEY, name TEXT NOT NULL, original_name TEXT NOT NULL DEFAULT '',
    overview TEXT NOT NULL DEFAULT '', first_air_date DATE, poster_path TEXT NOT NULL DEFAULT '',
    popularity DOUBLE PRECISION NOT NULL DEFAULT 0, raw_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS tv_episodes (
    tmdb_id BIGINT PRIMARY KEY, series_id BIGINT NOT NULL REFERENCES tv_series(tmdb_id) ON DELETE CASCADE,
    season_number INTEGER NOT NULL, episode_number INTEGER NOT NULL, name TEXT NOT NULL DEFAULT '',
    air_date DATE, raw_data JSONB NOT NULL DEFAULT '{}'::jsonb, updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(series_id, season_number, episode_number)
);
CREATE TABLE IF NOT EXISTS title_aliases (
    id BIGSERIAL PRIMARY KEY, media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    tmdb_id BIGINT NOT NULL, language_code TEXT NOT NULL DEFAULT '', alias TEXT NOT NULL,
    UNIQUE(media_type, tmdb_id, language_code, alias)
);
CREATE TABLE IF NOT EXISTS sync_runs (
    id BIGSERIAL PRIMARY KEY, started_at TIMESTAMPTZ NOT NULL DEFAULT now(), finished_at TIMESTAMPTZ,
    state TEXT NOT NULL CHECK (state IN ('running', 'success', 'failed')),
    processed INTEGER NOT NULL DEFAULT 0, error TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS sync_changes (
    id BIGSERIAL PRIMARY KEY, media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    tmdb_id BIGINT NOT NULL, action TEXT NOT NULL, changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS movies_title_search_idx ON movies USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS tv_series_name_search_idx ON tv_series USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS sync_changes_changed_at_idx ON sync_changes (changed_at DESC);
CREATE INDEX IF NOT EXISTS tv_episodes_series_idx ON tv_episodes (series_id, season_number, episode_number);
