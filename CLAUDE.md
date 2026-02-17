# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

CityQuiz — a quiz game backend. Single Go process with embedded SQLite (Turso go-libsql). No Redis, no distributed state.

## Commands

All commands run from `api/`.

```bash
go run ./cmd/server              # start server (default :8080, local.db)
go test ./...                    # all tests
go test -run TestName ./...      # single test
go build ./...                   # build check
go mod tidy                      # clean deps
```

If `TestHandleWSEcho` fails in a sandboxed environment (blocked socket bind), run non-socket packages:
```bash
go test ./cmd/server ./internal/config ./internal/database ./internal/migrations
```

## Environment Variables

| Var | Default | Notes |
|-----|---------|-------|
| `DB_PATH` | `local.db` | SQLite file path |
| `HTTP_ADDR` | `:8080` | Listen address |
| `LOG_LEVEL` | `INFO` | slog level |

## Architecture

**Flat idiomatic Go.** No layered "clean architecture." Handlers are closures in `server/`, not separate packages. No domain/model package unless interfaces earn their keep with multiple implementations.

```
api/
  cmd/server/main.go          — bootstrap: config → db → migrations → server → errgroup
  internal/
    config/                    — env-based config (caarlos0/env)
    database/                  — SQLite connection + PRAGMAs (WAL, busy_timeout, foreign_keys)
    migrations/                — goose v3, embedded SQL files (//go:embed *.sql)
    server/                    — chi router, middleware, all HTTP/WS handlers
```

Startup order: load config → open DB → run migrations → start HTTP server. Graceful shutdown via errgroup + signal.NotifyContext.

## Key Dependencies

- **go-libsql** (Turso) — SQLite driver, requires CGO. PRAGMAs must use `QueryContext` not `ExecContext` (driver quirk).
- **chi/v5** — router and middleware (RequestID, RealIP, structured logger, Recoverer).
- **goose/v3** — migrations from embedded SQL. Runs at startup automatically.
- **swaggest/openapi-go** — OpenAPI 3.0 spec generated from Go structs via reflector.
- **swaggest/swgui** — embedded Swagger UI v5 served at `/docs`.
- **nhooyr.io/websocket** — WebSocket support.

## Database

SQLite with WAL mode. All IDs are 16-byte random hex (`randomblob(16)`). Timestamps are ISO 8601 UTC. `:memory:` works for tests.

## Design Rules

- Split packages at ~800 lines, not before.
- Concrete types by default; interfaces only with a real second implementation.
- Keep OpenAPI spec in sync — it's generated from handler structs, so add response types at package level.
- SQLite is the only datastore. No external state infra unless explicitly requested.
