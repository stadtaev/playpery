#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

PIDS=()
cleanup() {
    trap - EXIT INT TERM
    echo ""
    echo "shutting down..."
    kill "${PIDS[@]}" 2>/dev/null
    wait 2>/dev/null
    exit 0
}
trap cleanup EXIT INT TERM

# Build frontend once so Go can serve it immediately.
echo "==> building frontend..."
(cd "$ROOT/web" && pnpm build) 2>&1 | sed 's/^/[build] /'

# Start Go server.
echo "==> starting api on :8080..."
(cd "$ROOT/api" && exec go run ./cmd/server) 2>&1 | sed 's/^/[api] /' &
PIDS+=($!)

# Start Vite dev server (hot reload, proxies /api to Go).
echo "==> starting vite on :5173..."
(cd "$ROOT/web" && exec pnpm dev) 2>&1 | sed 's/^/[web] /' &
PIDS+=($!)

echo ""
echo "  SPA (hot reload): http://localhost:5173/join/incas-2025"
echo "  Go  (full stack): http://localhost:8080/join/incas-2025"
echo "  API docs:         http://localhost:8080/docs"
echo ""

wait
