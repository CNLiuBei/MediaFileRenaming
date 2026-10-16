CREATE INDEX IF NOT EXISTS titles_title_prefix_idx
    ON titles (lower(title) text_pattern_ops);
CREATE INDEX IF NOT EXISTS titles_original_title_prefix_idx
    ON titles (lower(original_title) text_pattern_ops);
CREATE INDEX IF NOT EXISTS titles_english_title_prefix_idx
    ON titles (lower(english_title) text_pattern_ops);
