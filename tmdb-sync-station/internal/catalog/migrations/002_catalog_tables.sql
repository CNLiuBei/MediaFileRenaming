CREATE TABLE IF NOT EXISTS titles (
    tmdb_id BIGINT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('movie', 'tv')),
    title TEXT NOT NULL,
    original_title TEXT NOT NULL,
    english_title TEXT NOT NULL DEFAULT '',
    year INTEGER,
    overview TEXT NOT NULL DEFAULT '',
    popularity DOUBLE PRECISION NOT NULL DEFAULT 0,
    vote_average DOUBLE PRECISION NOT NULL DEFAULT 0,
    original_language TEXT NOT NULL DEFAULT '',
    genre_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    origin_countries JSONB NOT NULL DEFAULT '[]'::jsonb,
    director TEXT NOT NULL DEFAULT '',
    cast_members JSONB NOT NULL DEFAULT '[]'::jsonb,
    networks JSONB NOT NULL DEFAULT '[]'::jsonb,
    companies JSONB NOT NULL DEFAULT '[]'::jsonb,
    poster_path TEXT NOT NULL DEFAULT '',
    backdrop_path TEXT NOT NULL DEFAULT '',
    runtime INTEGER,
    vote_count INTEGER NOT NULL DEFAULT 0,
    imdb_id TEXT NOT NULL DEFAULT '',
    tagline TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    homepage TEXT NOT NULL DEFAULT '',
    last_air_date TEXT NOT NULL DEFAULT '',
    season_count INTEGER NOT NULL DEFAULT 0,
    episode_count INTEGER NOT NULL DEFAULT 0,
    in_production BOOLEAN NOT NULL DEFAULT false,
    show_type TEXT NOT NULL DEFAULT '',
    collection_id BIGINT,
    collection_name TEXT NOT NULL DEFAULT '',
    detail_version INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tmdb_id, kind)
);
CREATE INDEX IF NOT EXISTS titles_kind_year_idx ON titles (kind, year);

CREATE TABLE IF NOT EXISTS aliases (
    tmdb_id BIGINT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('movie', 'tv')),
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    PRIMARY KEY (tmdb_id, kind, name_key),
    FOREIGN KEY (tmdb_id, kind) REFERENCES titles(tmdb_id, kind) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS aliases_lookup_idx ON aliases (kind, name_key);

CREATE TABLE IF NOT EXISTS seasons (
    show_id BIGINT NOT NULL,
    season_number INTEGER NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    episode_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (show_id, season_number)
);

CREATE TABLE IF NOT EXISTS episodes (
    show_id BIGINT NOT NULL,
    season INTEGER NOT NULL,
    episode INTEGER NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (show_id, season, episode)
);

CREATE TABLE IF NOT EXISTS catalog_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
