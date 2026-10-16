CREATE TABLE IF NOT EXISTS sync_batches (
    id BIGSERIAL PRIMARY KEY,
    window_start DATE NOT NULL,
    window_end DATE NOT NULL,
    clipped BOOLEAN NOT NULL DEFAULT false,
    movie_page INTEGER NOT NULL DEFAULT 1,
    movie_total_pages INTEGER NOT NULL DEFAULT 1,
    tv_page INTEGER NOT NULL DEFAULT 1,
    tv_total_pages INTEGER NOT NULL DEFAULT 1,
    manifest_complete BOOLEAN NOT NULL DEFAULT false,
    completed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS sync_batches_one_open_idx
    ON sync_batches ((completed)) WHERE completed = false;

CREATE TABLE IF NOT EXISTS sync_batch_items (
    batch_id BIGINT NOT NULL REFERENCES sync_batches(id) ON DELETE CASCADE,
    media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    tmdb_id BIGINT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (batch_id, media_type, tmdb_id)
);

CREATE INDEX IF NOT EXISTS sync_batch_items_pending_idx
    ON sync_batch_items (batch_id, media_type, tmdb_id) WHERE completed = false;
