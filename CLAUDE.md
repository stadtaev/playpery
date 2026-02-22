# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

CityQuest — a location-based quest game SaaS. Players join teams via link/QR, walk through city landmarks, answer questions at each stage. Single Go process with embedded SQLite (Turso go-libsql), React SPA frontend. No Redis, no distributed state. Multi-tenant: each client gets its own SQLite database file.

## Commands

### Backend (run from `api/`)

```bash
go run ./cmd/server              # start server (default :8080, data/)
go test ./...                    # all tests
go test -run TestName ./...      # single test
go build ./...                   # build check
go mod tidy                      # clean deps
```

If `TestHandleWSEcho` fails in a sandboxed environment (blocked socket bind), run non-socket packages:
```bash
go test ./cmd/server ./internal/config ./internal/database
```

### Frontend (run from `web/`)

```bash
pnpm install                     # install deps
pnpm dev                         # dev server on :5173 (proxies /api to :8080)
pnpm build                       # production build to web/dist/
```

### Full stack

```bash
cd web && pnpm build
cd api && SPA_DIR=../web/dist go run ./cmd/server
# Open http://localhost:8080/join/demo/incas-2025
```

## Environment Variables

| Var | Default | Notes |
|-----|---------|-------|
| `DB_PATH` | `local.db` | SQLite file path; admin + client DBs sit in same directory |
| `HTTP_ADDR` | `:8080` | Listen address |
| `LOG_LEVEL` | `INFO` | slog level |
| `SPA_DIR` | `../web/dist` | Path to built SPA (`web/dist/`). If empty, no SPA serving. |

## Architecture

**Flat idiomatic Go.** No layered "clean architecture." Handlers are closures in `server/`, not separate packages. No domain/model package unless interfaces earn their keep with multiple implementations.

**Multi-tenant database layout:**
```
data/                          ← directory derived from DB_PATH
  _admin.db                    ← shared: admins, admin_sessions, clients
  demo.db                      ← per-client: scenarios, games, player_sessions
  {slug}.db                    ← one per client
```

```
api/
  cmd/server/main.go             — bootstrap: config → admin DB → registry → seed demo → server
  internal/
    config/                       — env-based config (caarlos0/env)
    database/                     — SQLite connection + PRAGMAs (WAL, busy_timeout, foreign_keys)
    server/
      server.go                   — http.Server setup, structured logger middleware
      routes.go                   — chi router, all route registration
      json.go                     — writeJSON, readJSON, writeError helpers
      auth.go                     — session token lookup (playerFromRequest)
      admin_auth.go               — admin session type + cookie name
      middleware.go               — clientMiddleware, adminAuthMiddleware, context helpers
      broker.go                   — in-process SSE pub/sub (mutex + map of teamID → channels)
      store.go                    — Store interface (client-scoped methods only)
      store_docs.go               — DocStore: JSONB-based Store implementation
      store_admin.go              — AdminAuth interface + AdminStore (shared admin DB)
      registry.go                 — Registry: maps client slugs to DocStore instances
      handle_team.go              — GET /api/{client}/teams/{joinToken}
      handle_join.go              — POST /api/{client}/join
      handle_game_state.go        — GET /api/{client}/game/state
      handle_answer.go            — POST /api/{client}/game/answer
      handle_events.go            — GET /api/{client}/game/events (SSE)
      handle_admin_login.go       — POST /api/admin/login, GET /api/admin/me, clients CRUD
      handle_admin_logout.go      — POST /api/admin/logout
      handle_admin_scenarios.go   — CRUD for /api/admin/clients/{client}/scenarios
      handle_admin_games.go       — CRUD for /api/admin/clients/{client}/games + nested teams
      spa.go                      — static file server + index.html fallback
      health.go                   — GET /healthz
      openapi.go                  — OpenAPI 3.0 spec generation
      wsecho.go                   — GET /ws/echo (WebSocket test)
web/
  src/
    types.ts                      — TS types matching API responses
    api.ts                        — fetch wrappers (client-scoped: /api/{client}/...)
    App.tsx                       — URL-based routing (no router library)
    JoinPage.tsx                  — team lookup → name input → join (/join/{client}/{token})
    GamePage.tsx                  — game state, clue, question, answer, timer
    useGameEvents.ts              — SSE hook (EventSource, client-aware)
    admin/
      adminTypes.ts               — TS types for admin API
      adminApi.ts                 — fetch wrappers (client-scoped: /api/admin/clients/{client}/...)
      AdminLoginPage.tsx          — email + password login
      AdminLayout.tsx             — auth check, nav, logout (client-aware breadcrumb)
      AdminClientsPage.tsx        — client list + create
      AdminScenariosPage.tsx      — scenario list + delete (per-client)
      AdminScenarioEditorPage.tsx — create/edit scenario with stages (per-client)
      AdminGamesPage.tsx          — game list + delete (per-client)
      AdminGameEditorPage.tsx     — create/edit game with teams section (per-client)
```

Startup order: load config → derive DB directory from DB_PATH → open admin DB → create Registry → pre-open existing clients → seed demo if first run → start HTTP server. Graceful shutdown via errgroup + signal.NotifyContext.

## API Endpoints

| Method | Path | Purpose | Auth |
|--------|------|---------|------|
| GET | `/healthz` | Health check | none |
| GET | `/openapi.json` | OpenAPI spec | none |
| GET | `/docs` | Swagger UI | none |
| GET | `/ws/echo` | WebSocket echo | none |
| GET | `/api/{client}/teams/{joinToken}` | Look up team before joining | none |
| POST | `/api/{client}/join` | Player joins team, gets session token | none |
| GET | `/api/{client}/game/state` | Full game state for player's team | Bearer |
| POST | `/api/{client}/game/answer` | Submit answer for current stage | Bearer |
| GET | `/api/{client}/game/events` | SSE stream for real-time updates | `?token=` |
| POST | `/api/admin/login` | Admin login (email+password → cookie) | none |
| POST | `/api/admin/logout` | Admin logout (clear session) | cookie |
| GET | `/api/admin/me` | Current admin info | cookie |
| GET | `/api/admin/clients` | List all clients | cookie |
| POST | `/api/admin/clients` | Create new client | cookie |
| GET | `/api/admin/clients/{client}/scenarios` | List all scenarios | cookie |
| POST | `/api/admin/clients/{client}/scenarios` | Create scenario with stages | cookie |
| GET | `/api/admin/clients/{client}/scenarios/{id}` | Get scenario detail | cookie |
| PUT | `/api/admin/clients/{client}/scenarios/{id}` | Update scenario | cookie |
| DELETE | `/api/admin/clients/{client}/scenarios/{id}` | Delete scenario (409 if games exist) | cookie |
| GET | `/api/admin/clients/{client}/games` | List all games | cookie |
| POST | `/api/admin/clients/{client}/games` | Create game | cookie |
| GET | `/api/admin/clients/{client}/games/{gameID}` | Get game with teams | cookie |
| PUT | `/api/admin/clients/{client}/games/{gameID}` | Update game | cookie |
| DELETE | `/api/admin/clients/{client}/games/{gameID}` | Delete game (409 if players exist) | cookie |
| GET | `/api/admin/clients/{client}/games/{gameID}/teams` | List teams for game | cookie |
| POST | `/api/admin/clients/{client}/games/{gameID}/teams` | Create team (auto-token) | cookie |
| PUT | `/api/admin/clients/{client}/games/{gameID}/teams/{teamID}` | Update team name/guide | cookie |
| DELETE | `/api/admin/clients/{client}/games/{gameID}/teams/{teamID}` | Delete team (409 if players) | cookie |

**Player auth:** session token (opaque hex). `Authorization: Bearer {token}` for REST, `?token=` query param for SSE.

**Admin auth:** `admin_session` HttpOnly cookie. Default credentials: `admin@playperu.com` / `changeme`.

## Key Dependencies

### Backend
- **go-libsql** (Turso) — SQLite driver, requires CGO. PRAGMAs must use `QueryContext` not `ExecContext` (driver quirk).
- **chi/v5** — router and middleware (RequestID, RealIP, structured logger, Recoverer).
- **swaggest/openapi-go** — OpenAPI 3.0 spec generated from Go structs via reflector.
- **swaggest/swgui** — embedded Swagger UI v5 served at `/docs`.
- **nhooyr.io/websocket** — WebSocket support.
- **golang.org/x/crypto/bcrypt** — admin password hashing.

### Frontend
- **Vite** — build tool, dev server with proxy.
- **React 19** + TypeScript.
- **pnpm** — fast, disk-efficient package manager.
- **Pico.css** (CDN) — minimal classless CSS framework.

## Database

**Per-client SQLite** with WAL mode. DocStore creates its own tables (JSONB schema evolution). Admin DB (`_admin.db`) stores admins, admin sessions, and client registry. All IDs are 16-byte random hex. Timestamps are ISO 8601 UTC. `:memory:` works for tests.

## Design Rules

- Split packages at ~800 lines, not before.
- Concrete types by default; interfaces only with a real second implementation (Store, AdminAuth).
- Keep OpenAPI spec in sync — it's generated from handler structs, so add response types at package level.
- SQLite is the only datastore. No external state infra unless explicitly requested.
- Timer check is lazy (computed on each request from `started_at + timer_minutes`). No background goroutines.
- SSE broker is in-process (no Redis pub/sub). Frontend re-fetches full state on every SSE event.
- Handlers get store from request context via `clientStore(r)`, not as closure parameters.
- Admin auth is enforced via `adminAuthMiddleware`, not per-handler checks.
