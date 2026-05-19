#!/usr/bin/env bash
# One-off new server bootstrap. Run from your local machine:
#   ./deploy/bootstrap.sh root@YOUR_SERVER_IP
set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: ./deploy/bootstrap.sh user@host"
    exit 1
fi

SERVER="$1"
DEPLOY_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "==> Ensuring rsync is installed on $SERVER..."
ssh "$SERVER" 'command -v rsync >/dev/null 2>&1 || { \
    if command -v apt-get >/dev/null 2>&1; then apt-get update && apt-get install -y rsync; \
    elif command -v dnf >/dev/null 2>&1; then dnf install -y rsync; \
    elif command -v yum >/dev/null 2>&1; then yum install -y rsync; \
    else echo "No supported package manager found; install rsync manually." >&2; exit 1; \
    fi; }'

echo "==> Uploading deploy files to $SERVER..."
rsync -avz --progress \
    "$DEPLOY_DIR/" \
    "$SERVER:/tmp/cityquest-deploy/"

echo "==> Running setup on $SERVER..."
ssh "$SERVER" 'bash /tmp/cityquest-deploy/setup.sh'

echo "==> Cleaning up temp files..."
ssh "$SERVER" 'rm -rf /tmp/cityquest-deploy'

echo ""
echo "=== Bootstrap complete ==="
echo ""
echo "Now deploy the app:"
echo "  ./deploy/deploy.sh $SERVER"
