-- +goose Up
CREATE VIRTUAL TABLE cue_index USING fts5(
    lecture_key UNINDEXED,
    lang UNINDEXED,
    start_ms UNINDEXED,
    text UNINDEXED,
    search_text,
    tokenize = 'unicode61 remove_diacritics 2'
);

-- +goose Down
DROP TABLE cue_index;
