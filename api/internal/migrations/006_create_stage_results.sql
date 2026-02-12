-- +goose Up
CREATE TABLE stage_results (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    game_id      TEXT NOT NULL REFERENCES games(id),
    team_id      TEXT NOT NULL REFERENCES teams(id),
    stage_number INTEGER NOT NULL,
    answer       TEXT,
    is_correct   INTEGER,  -- 0 or 1
    answered_at  TEXT
);

-- +goose Down
DROP TABLE stage_results;
