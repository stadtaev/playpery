#!/usr/bin/env bash
# Deploy to a Debian 13 (trixie, glibc 2.41) server from a host with newer glibc.
#
# Why this exists: the default deploy.sh builds locally. If your build host has
# a newer glibc than the server (e.g. Fedora 44 = glibc 2.43, server = 2.41),
# the binary fails to load with:
#   /lib/x86_64-linux-gnu/libm.so.6: version `GLIBC_2.43' not found
# This script builds the Go binary inside a Debian trixie Docker container so
# the binary's glibc requirement matches the server.
#
# Requirements (local):
#   - podman (or docker — set CONTAINER_CMD below)
#   - rsync, ssh
#
# Usage:
#   ./deploy/deploy-ruvds-debian13.sh root@SERVER_IP

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: ./deploy/deploy-ruvds-debian13.sh user@host"
    exit 1
fi

SERVER="$1"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Container runtime: podman by default, override with CONTAINER_CMD=docker.
CONTAINER_CMD="${CONTAINER_CMD:-podman}"

# Keep this tag in sync with api/go.mod's `go` directive.
# Bump to golang:1.26-trixie when go.mod is bumped.
BUILDER_IMAGE="docker.io/library/golang:1.25-trixie"

echo "==> Building frontend..."
cd "$REPO_ROOT/web"
pnpm build

echo "==> Building Go binary in $BUILDER_IMAGE (via $CONTAINER_CMD)..."
$CONTAINER_CMD run --rm \
    -v "$REPO_ROOT/api:/src:z" \
    -v "$REPO_ROOT:/out:z" \
    -v cityquest-gocache:/root/.cache/go-build \
    -v cityquest-gomod:/go/pkg/mod \
    -w /src \
    -e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=amd64 \
    "$BUILDER_IMAGE" \
    go build -o /out/cityquest ./cmd/server

echo "==> Uploading binary to $SERVER..."
rsync -avz --progress \
    "$REPO_ROOT/cityquest" \
    "$SERVER:/opt/cityquest/cityquest"

echo "==> Uploading SPA assets to $SERVER..."
rsync -avz --delete --progress \
    "$REPO_ROOT/web/dist/" \
    "$SERVER:/opt/cityquest/web/"

echo "==> Restarting service..."
ssh "$SERVER" 'if [ "$(id -u)" -eq 0 ]; then systemctl restart cityquest; else sudo systemctl restart cityquest; fi'

rm -f "$REPO_ROOT/cityquest"

echo ""
echo "=== Deploy complete ==="
