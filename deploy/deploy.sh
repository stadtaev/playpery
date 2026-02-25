#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: ./deploy/deploy.sh user@host"
    exit 1
fi

SERVER="$1"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Building frontend..."
cd "$REPO_ROOT/web"
pnpm build

echo "==> Building Go binary (linux/amd64)..."
cd "$REPO_ROOT/api"

# If you're already on linux/amd64, CGO cross-compilation just works.
# On macOS/ARM, you need a cross-compiler (e.g. zig cc, or build on the server).
# To build on the server instead, comment out the local build and uncomment the
# rsync of api/ + remote build block below.
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o "$REPO_ROOT/cityquest" ./cmd/server

echo "==> Uploading to $SERVER..."
rsync -avz --progress \
    "$REPO_ROOT/cityquest" \
    "$SERVER:/opt/cityquest/cityquest"

rsync -avz --delete --progress \
    "$REPO_ROOT/web/dist/" \
    "$SERVER:/opt/cityquest/web/"

echo "==> Restarting service..."
ssh "$SERVER" 'sudo systemctl restart cityquest'

rm -f "$REPO_ROOT/cityquest"

echo "==> Done! Service restarted on $SERVER"
