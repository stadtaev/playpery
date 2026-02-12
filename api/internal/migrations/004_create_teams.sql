-- +goose Up
CREATE TABLE teams (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    game_id    TEXT NOT NULL REFERENCES games(id),
    name       TEXT NOT NULL,
    join_token TEXT UNIQUE NOT NULL,
    guide_name TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE teams;
