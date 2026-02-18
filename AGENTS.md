# AGENTS.md

## Purpose
This file is the contributor/agent operating guide for this repository.

## Scope
- Repository: `playperu`
- Backend code: `api/`
- Frontend code: `web/`
- Runtime model: single Go API process + embedded SQLite + React SPA (no Redis)

## Repo Map
- `api/cmd/server/main.go`: server bootstrap
- `api/internal/config`: env config loading (`HTTP_ADDR`, `DB_PATH`, `LOG_LEVEL`, `SPA_DIR`)
- `api/internal/database`: SQLite connection + PRAGMA setup
- `api/internal/migrations`: embedded SQL migrations (001–007)
- `api/internal/server`: HTTP routes, handlers, SSE broker, SPA serving
  - `json.go`: shared JSON response/request helpers
  - `auth.go`: session token lookup from Bearer header or query param
  - `broker.go`: in-process SSE pub/sub (mutex + map of channels per team)
  - `handle_team.go`: GET `/api/teams/{joinToken}`
  - `handle_join.go`: POST `/api/join`
  - `handle_game_state.go`: GET `/api/game/state`
  - `handle_answer.go`: POST `/api/game/answer`
  - `handle_events.go`: GET `/api/game/events` (SSE)
  - `spa.go`: static file server with index.html fallback
  - `health.go`: GET `/healthz`
  - `openapi.go`: OpenAPI 3.0 spec
  - `wsecho.go`: WebSocket echo test endpoint
- `web/src/`: React SPA (Vite + TypeScript)
  - `types.ts`, `api.ts`, `App.tsx`, `JoinPage.tsx`, `GamePage.tsx`, `useGameEvents.ts`
- `docker-compose.dev.yml`: optional dev-services placeholder (currently empty)

## Commands
Run from repo root unless noted.

### Backend
- Build: `cd api && go build ./...`
- Test: `cd api && go test ./...`
- Run: `cd api && go run ./cmd/server`
- Tidy: `cd api && go mod tidy`
- Format: `gofmt -w api/`

### Frontend
- Install: `cd web && npm install`
- Dev: `cd web && npm run dev` (proxies `/api/*` to `:8080`)
- Build: `cd web && npm run build`

### Full stack
- `cd web && npm run build && cd ../api && SPA_DIR=../web/dist go run ./cmd/server`

Sandbox note:
- In restricted environments, `TestHandleWSEcho` may fail due to blocked local socket bind from `httptest.NewServer`.
- In that case, run non-socket packages explicitly:
  - `cd api && go test ./cmd/server ./internal/config ./internal/database ./internal/migrations`

## Environment Variables
- `DB_PATH` (default `local.db`): SQLite path
- `HTTP_ADDR` (default `:8080`)
- `LOG_LEVEL` (default `INFO`)
- `SPA_DIR` (default empty): path to built SPA directory. If set, serves static files with SPA fallback.

## API Endpoints
- `GET /healthz`: SQLite health check
- `GET /openapi.json`: OpenAPI 3.0 spec
- `GET /docs`: Swagger UI
- `GET /ws/echo`: WebSocket echo test endpoint
- `GET /api/teams/{joinToken}`: look up team by join token (no auth)
- `POST /api/join`: player joins team, returns session token (no auth)
- `GET /api/game/state`: full game state for player's team (Bearer token)
- `POST /api/game/answer`: submit answer for current stage (Bearer token)
- `GET /api/game/events`: SSE stream for real-time updates (`?token=` query param)

## Architecture Rules
- SQLite is the single datastore and source of truth.
- Keep dependency footprint minimal.
- Use concrete types by default; introduce interfaces only with a real second implementation.
- Keep state transitions server-authoritative.
- Timer is lazy (computed on request, no background goroutines).
- SSE broker is in-process. Frontend re-fetches full state on every event.
- Handlers are closures in `server/`, not in separate handler packages.
- Split packages at ~800 lines, not before.

## Database Rules
- Apply migrations at startup.
- Preserve SQLite pragmas in `database.Open`:
  - `journal_mode=WAL`
  - `busy_timeout=5000`
  - `foreign_keys=ON`
- Do not add external state infra (Redis, cache brokers) unless explicitly requested.
- All IDs are 16-byte random hex. Timestamps are ISO 8601 UTC.

## API Rules
- Health endpoint must reflect real runtime dependencies only.
- WebSocket handlers must remain context-aware and close connections correctly.
- Keep OpenAPI spec in sync with implemented routes and payloads.
- Avoid cross-file abstractions unless they reduce real complexity.
- Answer checking: `strings.EqualFold(TrimSpace(submitted), TrimSpace(correct))`.

## Frontend Rules
- No React Router — URL-based conditional rendering.
- No state management library — `useState` + `useEffect`.
- Pico.css via CDN for styling. No CSS framework build step.
- Session stored in `localStorage` (`session_token`).
- SSE via native `EventSource`. Re-fetch full state on every event.

## Editing Standards
- Keep changes small and local.
- Update tests/docs in the same change when behavior changes.
- Remove dead code; do not leave placeholder scaffolding.
- Prefer standard library and existing project dependencies.

## Done Criteria
A change is complete when all are true:
- Code compiles (`go build ./...` and `npm run build`).
- Relevant tests pass.
- `go.mod`/`go.sum` are clean.
- `CLAUDE.md` and `AGENTS.md` stay aligned with current code.
