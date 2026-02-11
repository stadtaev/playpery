# CityQuiz — Architecture & Development Guide

## Project Overview

CityQuiz is a SaaS platform for city-based team quiz games. Client companies purchase games and select scenarios. Teams of 3-7 people meet guides at starting locations, join via QR code, and navigate through city spots answering questions. The platform supports multiple concurrent games with real-time team tracking.

## Game Flow

1. **Client** purchases a game and selects a scenario from the catalog
2. **Admin** creates a game instance from the scenario, assigns teams
3. **Guide** (1 per team) meets their team at the starting spot and shows them a QR code
4. **Players** scan QR → join the React web app → enter their name → connected to their team room via WebSocket
5. **During the game**: players see their current stage, next place clue, and countdown timer. At each spot, a designated person asks the team a question. One player inputs the answer.
6. **Admin dashboard** shows all teams' progress in real-time. Admin can broadcast messages to all teams or specific teams.
7. **After game ends**: join tokens are invalidated, app shows "game over" screen, links stop working.

## Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Backend | **Go** (Golang) | Excellent concurrency model for WebSockets, simple deployment, fast |
| WebSocket | **gorilla/websocket** or **nhooyr/websocket** | Mature Go WebSocket libraries |
| Frontend | **React SPA** (PWA-capable) | Single codebase for player app, guide view, and admin dashboard |
| Database | **PostgreSQL** | Game configs, scenarios, results, persistent data |
| Cache/State | **Redis** | Live game state, pub/sub for real-time updates, session tokens |
| Reverse Proxy | **Caddy** | Auto TLS, WebSocket proxying, static file serving |
| Infra | **Hetzner VPS** + **Docker Compose** | Simple, cheap (~€10-12/mo), reliable |
| IaC | **Terraform** (Hetzner + Cloudflare providers) | Reproducible infrastructure |
| CDN/DNS | **Cloudflare** (free tier) | DNS, DDoS protection, CDN for static assets |

## Project Structure

```
cityquiz/
├── api/                      # Go backend
│   ├── cmd/server/           # main.go entrypoint
│   ├── internal/
│   │   ├── handler/          # HTTP + WebSocket handlers
│   │   ├── game/             # Game engine (state machine, answer checking, timer)
│   │   ├── ws/               # WebSocket hub, room management, broadcasting
│   │   ├── model/            # Domain types
│   │   ├── store/            # PostgreSQL repositories
│   │   └── cache/            # Redis state + pub/sub
│   ├── migrations/           # SQL migrations (golang-migrate or goose)
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── web/                      # React SPA
│   ├── src/
│   │   ├── views/
│   │   │   ├── player/       # Player game view
│   │   │   ├── guide/        # Guide team management view
│   │   │   ├── admin/        # Admin dashboard
│   │   │   └── join/         # QR join + name entry screen
│   │   ├── hooks/            # useWebSocket, useGameState, useTimer
│   │   ├── components/       # Shared UI components
│   │   └── lib/              # API client, WebSocket client
│   ├── package.json
│   └── Dockerfile
├── infra/                    # Terraform
│   ├── main.tf               # Provider config
│   ├── variables.tf           # Tokens, SSH keys, region
│   ├── server.tf              # Hetzner VPS instance
│   ├── network.tf             # Private network (future scaling)
│   ├── firewall.tf            # Ports 80, 443, 22 only
│   ├── dns.tf                 # Cloudflare DNS records
│   ├── volumes.tf             # Persistent storage for Postgres data
│   ├── cloud-init.yaml        # Bootstrap: install Docker, pull compose
│   ├── outputs.tf             # Server IP, connection strings
│   └── terraform.tfvars       # Secrets (gitignored)
├── docker-compose.yml         # Production compose
├── docker-compose.dev.yml     # Dev: Postgres + Redis only
├── Caddyfile                  # Reverse proxy config
├── .github/workflows/         # CI/CD (later)
├── .gitignore
├── CLAUDE.md                  # This file
└── README.md
```

## Architecture

### Real-Time Communication

WebSocket rooms are the core abstraction:

```
Room: game:{gameId}            ← Admin joins this; sees all teams' progress
Room: team:{gameId}:{teamId}   ← All players + guide on a team
```

- **Team isolation**: Players only join their team room. The server NEVER sends cross-team data to player sockets.
- **Admin room**: Admin socket joins the game-level room and receives aggregated progress from all teams.
- **Broadcast**: Admin can push messages to all teams (via game room fan-out) or a specific team room.

### Game State (Redis)

Live game state is stored in Redis per team. This is the source of truth during an active game:

```json
{
  "gameId": "g_123",
  "teamId": "t_456",
  "currentStage": 3,
  "stages": [
    { "stageNumber": 1, "status": "completed", "answer": "1902", "correct": true, "answeredAt": "..." },
    { "stageNumber": 2, "status": "completed", "answer": "Charles IV", "correct": false, "answeredAt": "..." },
    { "stageNumber": 3, "status": "active", "arrivedAt": null }
  ],
  "startedAt": "2026-02-11T10:00:00Z",
  "timerEndsAt": "2026-02-11T12:00:00Z"
}
```

When state changes → server pushes update to the team room via WebSocket. Every player sees the same state instantly.

### Join Flow (QR Code)

Best practice: **unique short-lived join links delivered as QR codes**.

1. Admin creates a game → system generates a **join token per team** (random string, e.g., `X7k9mP`)
2. Join URL: `https://app.cityquiz.io/play/{joinToken}`
3. Guide displays this as a QR code on their phone
4. Player scans → React SPA loads → player enters their name → WebSocket connects to team room
5. Token maps to game + team in Redis: `joinToken → { gameId, teamId, role }`
6. **No accounts, no passwords.** The join token IS the session.
7. After game ends → token is deleted from Redis → URL returns "game over" screen.

### Answer Submission

- Any player can submit an answer (or restrict to one — configurable per game)
- `POST /api/games/:gameId/teams/:teamId/stages/:stageNumber/answer` with `{ answer: "..." }`
- Server validates, checks against correct answer, updates Redis state
- Emits state update to team room (all players see result)
- Emits progress update to admin room
- Next clue is revealed to the team

### Timer

**Server-authoritative timer. Never trust client clocks.**

- Server stores `startedAt` and `timerEndsAt` timestamps
- Client receives these values and renders a local countdown
- On WebSocket reconnect, client re-syncs from server state
- Server validates time on answer submissions

### Admin Dashboard

- Real-time view of all teams: current stage, progress, time elapsed
- Broadcast message to all teams or a specific team (WebSocket push → toast/modal on player phones)
- Controls: pause timer, skip a stage, override scores, end game early

## Database Schema (PostgreSQL)

```sql
-- Client companies who purchase games
CREATE TABLE clients (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    email       TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Scenario templates (reusable game blueprints)
CREATE TABLE scenarios (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    city        TEXT NOT NULL,
    description TEXT,
    stages      JSONB NOT NULL,  -- array of { stageNumber, location, clue, question, correctAnswer, lat, lng }
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Game instances (one per booking)
CREATE TABLE games (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id  UUID REFERENCES scenarios(id),
    client_id    UUID REFERENCES clients(id),
    status       TEXT NOT NULL DEFAULT 'draft',  -- draft, active, paused, ended
    scheduled_at TIMESTAMPTZ,
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ,
    timer_minutes INT DEFAULT 120,
    created_at   TIMESTAMPTZ DEFAULT now()
);

-- Teams within a game
CREATE TABLE teams (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id     UUID REFERENCES games(id),
    name        TEXT NOT NULL,
    join_token  TEXT UNIQUE NOT NULL,  -- random token for QR join link
    guide_name  TEXT,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Players (ephemeral, created on join)
CREATE TABLE players (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID REFERENCES teams(id),
    name        TEXT NOT NULL,
    session_id  TEXT UNIQUE NOT NULL,  -- maps to WebSocket connection
    joined_at   TIMESTAMPTZ DEFAULT now()
);

-- Answer history (persisted from Redis after game)
CREATE TABLE stage_results (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id      UUID REFERENCES games(id),
    team_id      UUID REFERENCES teams(id),
    stage_number INT NOT NULL,
    answer       TEXT,
    is_correct   BOOLEAN,
    answered_at  TIMESTAMPTZ
);
```

## Infrastructure

### Hetzner VPS Setup

- **Server**: CX31 (4 vCPU, 8GB RAM, 80GB SSD) — ~€8/month
- **Volume**: 20GB attached volume for Postgres data (survives server recreation)
- **Firewall**: Ports 80 (HTTP), 443 (HTTPS), 22 (SSH — locked to admin IP)
- **Backups**: Hetzner snapshots + nightly `pg_dump` to Hetzner Object Storage

### Docker Compose (Production)

```yaml
services:
  api:
    build: ./api
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://cityquiz:secret@postgres:5432/cityquiz
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16
    volumes:
      - /mnt/data/postgres:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=cityquiz
      - POSTGRES_USER=cityquiz
      - POSTGRES_PASSWORD=secret

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data

  caddy:
    image: caddy:2
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - ./web/dist:/srv/web
      - caddy-data:/data

volumes:
  redis-data:
  caddy-data:
```

### Docker Compose (Development)

Only run dependencies locally. Run Go API and React dev server natively for hot reload:

```yaml
services:
  postgres:
    image: postgres:16
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=cityquiz
      - POSTGRES_USER=cityquiz
      - POSTGRES_PASSWORD=secret

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Terraform Resources

Key resources managed by Terraform:
- `hcloud_server` — VPS instance with cloud-init bootstrap
- `hcloud_volume` + `hcloud_volume_attachment` — persistent Postgres storage
- `hcloud_firewall` — port rules
- `hcloud_ssh_key` — deploy key
- `cloudflare_record` — DNS A record pointing to VPS IP (proxied)

Terraform state stored in S3-compatible backend (Hetzner Object Storage or Terraform Cloud free tier).

### Deployment

Simple SSH-based deploy (or GitHub Action later):

```bash
ssh cityquiz "cd /opt/cityquiz && docker compose pull api && docker compose up -d api"
```

### Scaling Path (when needed, not now)

1. Move Postgres to Hetzner Managed Database (~€15/mo) for automated backups/failover
2. Multiple game types → separate Go services in same Docker Compose
3. Geographic expansion → additional VPS in another Hetzner region
4. If orchestration needed → k3s or Coolify on top of existing setup

## Security (Lightweight)

| Concern | Solution |
|---------|----------|
| Teams can't see other teams | WebSocket rooms — server never sends cross-team data to player sockets |
| App stops after game ends | Invalidate join tokens in Redis; game status → "ended"; SPA shows game-over screen |
| No auth needed for players | Join token = session key, stored in Redis with TTL |
| Prevent token reuse | Tokens bound to gameId + game status check on every request |

## Build Order (MVP)

Build and validate in this sequence:

1. **Go API skeleton** — health check endpoint, WebSocket echo test, Redis + Postgres connection
2. **Database schema + migrations** — use golang-migrate or goose
3. **Scenario + Game CRUD** — REST endpoints for admin to create/manage scenarios and games
4. **Join flow** — token generation, QR code rendering, WebSocket room assignment on join
5. **Game engine** — stage progression state machine, answer checking, server-authoritative timer
6. **React player app** — join screen → game view (current stage, clue, answer input, timer)
7. **Guide view** — same React app, different role; sees team roster, can trigger "arrived at spot"
8. **Admin dashboard** — real-time team progress, broadcast messages, game controls
9. **Post-game** — results screen, token invalidation, persist Redis state to Postgres
10. **Terraform + deploy** — IaC for Hetzner + Cloudflare, Docker Compose production setup

## Code Style & Conventions

- Go: follow standard Go project layout; use `internal/` for non-exported packages
- Use context.Context for request-scoped values and cancellation
- Prefer table-driven tests in Go
- React: functional components with hooks; no class components
- Use TypeScript for the React frontend
- Migrations: numbered SQL files, never modify existing migrations
- Environment config via environment variables (12-factor)
- All times in UTC, stored as TIMESTAMPTZ in Postgres

## Key Libraries (Go)

- HTTP router: `chi` or `gin` (lightweight, standard)
- WebSocket: `nhooyr.io/websocket` (modern, context-aware) or `gorilla/websocket`
- PostgreSQL: `pgx/v5` (native Go Postgres driver, no ORM)
- Redis: `redis/go-redis/v9`
- Migrations: `golang-migrate/migrate`
- Config: `caarlos0/env` or `kelseyhightower/envconfig`
- QR generation: `skip2/go-qrcode`
