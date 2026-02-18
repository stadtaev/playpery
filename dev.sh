#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

cleanup() {
    echo ""
    echo "shutting down..."
    kill 0 2>/dev/null
    wait 2>/dev/null
}
trap cleanup EXIT INT TERM

# Build frontend once so Go can serve it immediately.
echo "==> building frontend..."
(cd "$ROOT/web" && pnpm build) 2>&1 | sed 's/^/[build] /'

# Start Go server.
echo "==> starting api on :8080..."
(cd "$ROOT/api" && go run ./cmd/server) 2>&1 | sed 's/^/[api] /' &

# Start Vite dev server (hot reload, proxies /api to Go).
echo "==> starting vite on :5173..."
(cd "$ROOT/web" && pnpm dev) 2>&1 | sed 's/^/[web] /' &

echo ""
echo "  SPA (hot reload): http://localhost:5173/join/incas-2025"
echo "  Go  (full stack): http://localhost:8080/join/incas-2025"
echo "  API docs:         http://localhost:8080/docs"
echo ""

wait
