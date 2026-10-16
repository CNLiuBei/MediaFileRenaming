CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS titles_search_trgm_idx
    ON titles USING GIN ((title || ' ' || original_title || ' ' || english_title) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS aliases_name_trgm_idx
    ON aliases USING GIN (name gin_trgm_ops);
