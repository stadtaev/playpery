# CityQuest

**Turn any city into a live quest adventure.**

CityQuest is a location-based multiplayer quest game platform built for tourism operators, team-building companies, and event organizers. Teams of players walk through real city landmarks, solve clues, and answer questions at each location — racing against the clock and each other.

Think of it as a digital scavenger hunt, but the venue is an entire city.

## The Idea

Cities are full of history hiding in plain sight. Tourists walk past centuries-old fountains, churches with underground catacombs, and streets where liberators once marched — often without knowing the stories behind them. Traditional guided tours work, but they're passive. People remember what they *do*, not what they're told.

CityQuest flips the script. Instead of following a guide and listening, players get clues that lead them to landmarks. At each stop, they answer a question about what they find. Get it right, unlock the next stage. Get it wrong, keep trying. The whole thing is timed, and every team member can contribute from their own phone.

The platform is designed as a SaaS — tourism companies anywhere in the world can create their own scenarios with custom routes, questions, and branding. One operator might run a walking tour through a historic district. Another might build a team-building event through city streets. A school might use it for a history field trip. Same engine, different content.

### Why It Works

- **Active learning beats passive tours.** Players engage with the city instead of just looking at it. They notice architectural details, read plaques, and actually remember what they learned.
- **Social by design.** Teams collaborate in real-time. Multiple people on the same team see live updates as teammates answer questions. It's competitive and cooperative at the same time.
- **Zero app install.** Players join via a link or QR code on their phone's browser. No app store, no downloads, no friction. Show up, scan, play.
- **Scales without guides.** Once a scenario is created, it can run for hundreds of teams simultaneously with zero staff. The city *is* the venue, the phone *is* the guide.
- **Works anywhere.** Any city with interesting landmarks can have a CityQuest scenario. The content is what makes it local — the tech is universal.

### How a Game Works

1. **An operator creates a scenario** — a sequence of stages, each tied to a real-world location. Every stage has a clue (to get players to the right spot), a question (about what they'll find there), and a correct answer.

2. **The operator creates a game** from that scenario, optionally enables timers, and generates teams with unique join links/QR codes.

3. **Players scan the QR code** on their phone, enter their name, and join their team. No account creation, no app download.

4. **The game begins.** Each team sees their first clue. They walk to the location, find the answer, and submit it.

5. **Real-time updates** keep the whole team in sync. When a teammate answers, everyone's screen updates instantly via Server-Sent Events. New players joining mid-game see the current state immediately.

6. **The game ends** when the timer runs out or all stages are completed. Teams see their final score.

### Answering Questions

Teams progress through stages in order. At each stage, the team gets **one attempt** to answer:

- **Correct answer** — the team earns a point and advances to the next stage.
- **Wrong answer** — the correct answer is revealed, no point is earned, and the team advances to the next stage anyway.

Answers are compared case-insensitively with whitespace trimmed. There are no retries — every stage is answered exactly once.

### Scoring

A team's score is the number of correctly answered stages out of the total. At game end, the summary shows "X of Y answered correctly." The completed stages list shows each stage color-coded as correct (green) or incorrect (red).

### Game Types

**Standard game** — all players join through the team's regular join link and have the same role.

**Supervised game** — when a game is marked as supervised, each team gets two links: a regular join link and a supervisor link. Players who join via the supervisor link are assigned the "supervisor" role; all others are "player." The role is stored on the session and visible in admin views. (Supervisor-specific permissions are reserved for future use.)

### Timers

Games support two optional timers, enabled together via the "Enable timer" setting:

- **Game timer** — a total time limit for the entire game (default 120 minutes). When the game timer expires, the game status changes to "ended" and no more answers are accepted.
- **Stage timer** — a per-stage countdown (default 10 minutes) that resets when a new stage begins. This is a client-side display to create urgency; it does not auto-advance the stage server-side.

Timer expiration is checked lazily on each API request. The frontend displays both countdowns and highlights them in red when under 60 seconds.

### Game Lifecycle

| Status | Meaning |
|--------|---------|
| `draft` | Game created, not yet playable. Teams can be added and configured. |
| `active` | Game is live. Players can join teams and answer questions. |
| `paused` | Temporarily halted. Players see "game is not active." |
| `ended` | Game is over (manually ended or timer expired). Scores are final. |

Transitioning to "active" records the start time. Transitioning to "ended" records the end time. Resetting to "draft" clears both timestamps.

## Tech Stack

CityQuest is deliberately simple. One Go binary, one SQLite file, one React SPA. No Kubernetes, no microservices, no Redis, no message queues. A single $5/month VPS can run it for thousands of concurrent players.

- **Backend:** Go with [chi](https://github.com/go-chi/chi) router, embedded SQLite via [Turso go-libsql](https://github.com/tursodatabase/go-libsql)
- **Frontend:** React + TypeScript, built with [Vite](https://vite.dev), styled with [Pico.css](https://picocss.com)
- **Real-time:** Server-Sent Events (SSE) with an in-process pub/sub broker
- **Auth:** Opaque session tokens (no JWT, no OAuth — players don't have accounts)
- **API docs:** Auto-generated OpenAPI 3.0 spec with embedded Swagger UI at `/docs`

### Architecture

```
Browser (React SPA)
    ↕ HTTP + SSE
Go server (:8080)
    ↕ SQL
SQLite (local.db)
```

That's it. The Go server handles everything: API requests, SSE streaming, and serving the built SPA as static files. In development, Vite runs on `:5173` and proxies API calls to the Go server on `:8080`.

## Getting Started

### Prerequisites

- Go 1.21+ (with CGO enabled — required by the SQLite driver)
- Node.js 18+ (with corepack enabled)
- pnpm (`corepack enable && corepack prepare pnpm@latest --activate`)

### Quick Start

```bash
# Build the frontend
cd web
pnpm install
pnpm build
cd ..

# Start the server (serves API + SPA)
cd api
SPA_DIR=../web/dist go run ./cmd/server
```

Open http://localhost:8080/join/demo/incas-2025 in your browser and play through the demo.

### Development Mode

Run the backend and frontend separately for hot reload:

```bash
# Terminal 1: Go server
cd api
go run ./cmd/server

# Terminal 2: Vite dev server (proxies /api to :8080)
cd web
pnpm dev
```

Open http://localhost:5173/join/demo/incas-2025.

### Running Tests

```bash
cd api
go test ./...
```

## API Reference

Interactive API docs are available at `/docs` when the server is running.

| Method | Path | Purpose | Auth |
|--------|------|---------|------|
| `GET` | `/api/{client}/teams/{joinToken}` | Look up team before joining | none |
| `POST` | `/api/{client}/join` | Player joins team, gets session token | none |
| `GET` | `/api/{client}/game/state` | Full game state for player's team | Bearer token |
| `POST` | `/api/{client}/game/answer` | Submit answer for current stage | Bearer token |
| `GET` | `/api/{client}/game/events` | SSE stream for real-time updates | `?token=` query |
| `GET` | `/healthz` | Health check | none |

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8080` | Server listen address |
| `DB_PATH` | `local.db` | SQLite database file path |
| `LOG_LEVEL` | `INFO` | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`) |
| `SPA_DIR` | _(empty)_ | Path to built frontend. If empty, only the API is served. |

## Project Structure

```
api/                            # Go backend
  cmd/server/main.go            # Entry point
  internal/
    config/                     # Environment-based config
    database/                   # SQLite connection + pragmas
    server/                     # HTTP handlers, routes, SSE broker
web/                            # React frontend
  src/
    App.tsx                     # URL-based routing
    JoinPage.tsx                # Team join flow
    GamePage.tsx                # Main game interface
    useGameEvents.ts            # SSE hook
    api.ts                      # API client
    types.ts                    # TypeScript types
```

## Deployment

CityQuest deploys as a single Go binary behind Caddy (automatic HTTPS) on any Linux VPS. Two scripts handle everything.

### Prerequisites

- A Linux server (e.g. Hetzner) with SSH access as root
- Your domain's DNS A record pointing to the server IP

### First-time server setup

Run once from your local machine to install Caddy, create the systemd service, and prepare directories:

```bash
./deploy/bootstrap.sh root@YOUR_SERVER_IP
```

This installs Caddy, creates a `cityquest` system user, sets up `/opt/cityquest/`, and enables the systemd service.

### Deploy (first time and every update)

```bash
./deploy/deploy.sh root@YOUR_SERVER_IP
```

This builds the frontend, cross-compiles the Go binary for linux/amd64, uploads both to the server, and restarts the service. Run it every time you want to ship a new version.

### What's running on the server

```
Caddy (:443) → reverse proxy → cityquest (:8080)
```

- Caddy handles TLS (auto Let's Encrypt) and proxies to the Go server
- The Go binary at `/opt/cityquest/cityquest` serves the API and the SPA from `/opt/cityquest/web/`
- SQLite databases live in `/opt/cityquest/data/`

### Useful commands

```bash
ssh root@SERVER 'systemctl status cityquest'     # check service status
ssh root@SERVER 'journalctl -u cityquest -f'     # live logs
ssh root@SERVER 'systemctl restart cityquest'     # manual restart
```

## License

[Business Source License 1.1](LICENSE) — you can read, fork, and modify the code, but you can't use it commercially without a license from us. On 2029-02-26 it converts to GPL v2.
