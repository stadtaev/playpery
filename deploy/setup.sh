#!/usr/bin/env bash
# One-time server setup for CityQuest.
# Run on the server: bash setup.sh
set -euo pipefail

echo "==> Creating cityquest system user..."
if ! id cityquest &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin cityquest
fi

echo "==> Creating directories..."
mkdir -p /opt/cityquest/{data,web}
chown -R cityquest:cityquest /opt/cityquest

echo "==> Installing Caddy..."
apt-get update -qq
apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt-get update -qq
apt-get install -y -qq caddy

echo "==> Installing systemd service..."
cp "$(dirname "$0")/cityquest.service" /etc/systemd/system/cityquest.service
systemctl daemon-reload
systemctl enable cityquest

echo "==> Installing Caddyfile..."
cp "$(dirname "$0")/Caddyfile" /etc/caddy/Caddyfile
systemctl reload caddy

echo ""
echo "=== Setup complete ==="
