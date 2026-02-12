-- +goose Up
CREATE TABLE games (
    id            TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    scenario_id   TEXT NOT NULL REFERENCES scenarios(id),
    client_id     TEXT NOT NULL REFERENCES clients(id),
    status        TEXT NOT NULL DEFAULT 'draft',  -- draft, active, paused, ended
    scheduled_at  TEXT,
    started_at    TEXT,
    ended_at      TEXT,
    timer_minutes INTEGER NOT NULL DEFAULT 120,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE games;
