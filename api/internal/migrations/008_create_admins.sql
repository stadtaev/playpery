-- +goose Up
CREATE TABLE admins (
    id            TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE admin_sessions (
    id         TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    admin_id   TEXT NOT NULL REFERENCES admins(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE admin_sessions;
DROP TABLE admins;
