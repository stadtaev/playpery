-- +goose Up
CREATE TABLE players (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    team_id    TEXT NOT NULL REFERENCES teams(id),
    name       TEXT NOT NULL,
    session_id TEXT UNIQUE NOT NULL,
    joined_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE players;
