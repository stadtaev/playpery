-- +goose Up
CREATE TABLE scenarios (
    id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name        TEXT NOT NULL,
    city        TEXT NOT NULL,
    description TEXT,
    stages      TEXT NOT NULL,  -- JSON array: [{stageNumber, location, clue, question, correctAnswer, lat, lng}]
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE scenarios;
