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
