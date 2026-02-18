# CityQuiz by PlayPeru

**Turn any city into a live quiz adventure.**

CityQuiz is a location-based multiplayer quiz game platform built for tourism operators, team-building companies, and event organizers. Teams of players walk through real city landmarks, solve clues, and answer questions at each location — racing against the clock and each other.

Think of it as a digital scavenger hunt meets pub quiz, but the pub is an entire city.

## The Idea

Peru is full of cities with rich history hiding in plain sight. Tourists walk past 400-year-old fountains, colonial churches with underground catacombs, and streets where liberators once marched — often without knowing the stories behind them. Traditional guided tours work, but they're passive. People remember what they *do*, not what they're told.

CityQuiz flips the script. Instead of following a guide and listening, players get clues that lead them to landmarks. At each stop, they answer a question about what they find. Get it right, unlock the next stage. Get it wrong, keep trying. The whole thing is timed, and every team member can contribute from their own phone.

The platform is designed as a SaaS — tourism companies in Lima, Cusco, Arequipa (or anywhere in the world) can create their own scenarios with custom routes, questions, and branding. One operator might run a "Lima Centro Historico" game for tourists. Another might build a team-building event through the streets of Miraflores. A school might use it for a history field trip. Same engine, different content.

### Why It Works

- **Active learning beats passive tours.** Players engage with the city instead of just looking at it. They notice architectural details, read plaques, and actually remember what they learned.
- **Social by design.** Teams collaborate in real-time. Multiple people on the same team see live updates as teammates answer questions. It's competitive and cooperative at the same time.
- **Zero app install.** Players join via a link or QR code on their phone's browser. No app store, no downloads, no friction. Show up, scan, play.
- **Scales without guides.** Once a scenario is created, it can run for hundreds of teams simultaneously with zero staff. The city *is* the venue, the phone *is* the guide.
- **Works anywhere.** The platform isn't Peru-specific despite the name. Any city with interesting landmarks can have a CityQuiz scenario. The content is what makes it local — the tech is universal.

### How a Game Works

1. **An operator creates a scenario** — a sequence of stages, each tied to a real-world location. Every stage has a clue (to get players to the right spot), a question (about what they'll find there), and a correct answer.

2. **The operator creates a game** from that scenario, sets a timer (e.g. 2 hours), and generates teams with unique join links/QR codes.

3. **Players scan the QR code** on their phone, enter their name, and join their team. No account creation, no app download.

4. **The game begins.** Each team sees their first clue. They walk to the location, find the answer, and submit it. Correct answer → next stage unlocked. Wrong answer → try again (the correct answer is logged in the browser console for debugging during development).

5. **Real-time updates** keep the whole team in sync. When a teammate answers correctly, everyone's screen updates instantly via Server-Sent Events. New players joining mid-game see the current state immediately.

6. **The game ends** when the timer runs out or all stages are completed. Teams see their final score — how many stages they completed and how long it took.

### The Demo Scenario: Lima Centro Historico

The repository ships with a built-in demo scenario that takes players through four iconic landmarks in Lima's historic center:

| Stage | Location | Question |
|-------|----------|----------|
| 1 | **Plaza Mayor** | What year was the fountain in Plaza Mayor built? |
| 2 | **Iglesia de San Francisco** | What are the underground tunnels beneath San Francisco called? |
| 3 | **Jiron de la Union** | Which liberator has a statue on Jiron de la Union? |
| 4 | **Parque de la Muralla** | What century were the original city walls built in? |

Two teams are pre-configured: **Los Incas** (join token: `incas-2025`) and **Los Condores** (join token: `condores-2025`).

## Tech Stack

CityQuiz is deliberately simple. One Go binary, one SQLite file, one React SPA. No Kubernetes, no microservices, no Redis, no message queues. A single $5/month VPS can run it for thousands of concurrent players.

- **Backend:** Go with [chi](https://github.com/go-chi/chi) router, embedded SQLite via [Turso go-libsql](https://github.com/tursodatabase/go-libsql), automatic migrations with [goose](https://github.com/pressly/goose)
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
# Clone the repo
git clone https://github.com/playperu/cityquiz.git
cd cityquiz

# Build the frontend
cd web
pnpm install
pnpm build
cd ..

# Start the server (serves API + SPA)
cd api
SPA_DIR=../web/dist go run ./cmd/server
```

Open http://localhost:8080/join/incas-2025 in your browser and play through the demo.

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

Open http://localhost:5173/join/incas-2025.

### Running Tests

```bash
cd api
go test ./...
```

## API Reference

Interactive API docs are available at `/docs` when the server is running.

| Method | Path | Purpose | Auth |
|--------|------|---------|------|
| `GET` | `/api/teams/{joinToken}` | Look up team before joining | none |
| `POST` | `/api/join` | Player joins team, gets session token | none |
| `GET` | `/api/game/state` | Full game state for player's team | Bearer token |
| `POST` | `/api/game/answer` | Submit answer for current stage | Bearer token |
| `GET` | `/api/game/events` | SSE stream for real-time updates | `?token=` query |
| `GET` | `/healthz` | Health check | none |

### Example Flow

```bash
# Look up a team
curl localhost:8080/api/teams/incas-2025

# Join the team
curl -X POST localhost:8080/api/join \
  -H 'Content-Type: application/json' \
  -d '{"joinToken":"incas-2025","playerName":"Maria"}'
# → {"token":"abc123...","playerId":"...","teamId":"...","teamName":"Los Incas"}

# Get game state
curl -H 'Authorization: Bearer abc123...' localhost:8080/api/game/state

# Submit an answer
curl -X POST localhost:8080/api/game/answer \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer abc123...' \
  -d '{"answer":"1651"}'

# Listen for real-time events
curl -N 'localhost:8080/api/game/events?token=abc123...'
```

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
playperu/
  api/                            # Go backend
    cmd/server/main.go            # Entry point
    internal/
      config/                     # Environment-based config
      database/                   # SQLite connection + pragmas
      migrations/                 # SQL migration files (001–007)
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

## License

TBD
