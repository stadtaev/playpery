# AGENTS.md

## Purpose
This file is the contributor/agent operating guide for this repository.

## Scope
- Repository: `playperu`
- Backend code: `api/`
- Runtime model: single Go API process + embedded SQLite (no Redis)

## Repo Map
- `api/cmd/server/main.go`: server bootstrap
- `api/internal/config`: env config loading
- `api/internal/database`: SQLite connection + PRAGMA setup
- `api/internal/migrations`: embedded SQL migrations
- `api/internal/server`: HTTP routes and handlers
- `docker-compose.dev.yml`: optional dev-services placeholder (currently empty)

## Commands
Run from repo root unless noted.

- Format Go:
  - `gofmt -w api/cmd/server/main.go api/internal/config/config.go api/internal/database/database.go api/internal/migrations/migrations.go api/internal/server/*.go`
- Tidy modules:
  - `cd api && go mod tidy`
- Run API:
  - `cd api && DB_PATH=./cityquiz.db go run ./cmd/server`
- Run tests:
  - `cd api && go test ./...`

Sandbox note:
- In restricted environments, `TestHandleWSEcho` may fail due blocked local socket bind from `httptest.NewServer`.
- In that case, run non-socket packages explicitly:
  - `cd api && go test ./cmd/server ./internal/config ./internal/database ./internal/migrations`

## Environment Variables
- `DB_PATH` (required): SQLite path
- `HTTP_ADDR` (default `:8080`)
- `LOG_LEVEL` (default `INFO`)

## API Endpoints
- `GET /healthz`: SQLite health check
- `GET /openapi.json`: OpenAPI 3.1 spec
- `GET /docs`: Swagger UI
- `GET /ws/echo`: WebSocket echo test endpoint

## Architecture Rules
- SQLite is the single datastore and source of truth.
- Keep dependency footprint minimal.
- Use concrete types by default; introduce interfaces only with a real second implementation.
- Keep state transitions server-authoritative.

## Database Rules
- Apply migrations at startup.
- Preserve SQLite pragmas in `database.Open`:
  - `journal_mode=WAL`
  - `busy_timeout=5000`
  - `foreign_keys=ON`
- Do not add external state infra (Redis, cache brokers) unless explicitly requested.

## API Rules
- Health endpoint must reflect real runtime dependencies only.
- WebSocket handlers must remain context-aware and close connections correctly.
- Keep OpenAPI spec in sync with implemented routes and payloads.
- Avoid cross-file abstractions unless they reduce real complexity.

## Editing Standards
- Keep changes small and local.
- Update tests/docs in the same change when behavior changes.
- Remove dead code; do not leave placeholder scaffolding.
- Prefer standard library and existing project dependencies.

## Done Criteria
A change is complete when all are true:
- Code compiles.
- Relevant tests pass.
- `go.mod`/`go.sum` are clean.
- `AGENTS.md` stays aligned with current code.
