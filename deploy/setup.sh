#!/usr/bin/env bash
# One-time server setup for CityQuest.
# Run on the server: bash setup.sh
set -euo pipefail

echo "==> Creating cityquest system user..."
if ! id cityquest &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin cityquest
fi

echo "==> Creating directories..."
mkdir -p /opt/cityquest/{data,web,tls}

echo "==> Generating self-signed TLS cert for backend..."
if [ ! -f /opt/cityquest/tls/cert.pem ]; then
    openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
        -keyout /opt/cityquest/tls/key.pem \
        -out /opt/cityquest/tls/cert.pem \
        -days 3650 -nodes \
        -subj "/CN=localhost"
    chmod 600 /opt/cityquest/tls/key.pem
fi

chown -R cityquest:cityquest /opt/cityquest

echo "==> Installing Caddy..."
if ! command -v caddy &>/dev/null; then
    apt-get update -qq
    apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https curl
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --batch --yes --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
    apt-get update -qq
    apt-get install -y -qq caddy
else
    echo "    Caddy already installed, skipping"
fi

echo "==> Installing systemd service..."
cp "$(dirname "$0")/cityquest.service" /etc/systemd/system/cityquest.service
systemctl daemon-reload
systemctl enable cityquest

echo "==> Installing Caddyfile..."
cp "$(dirname "$0")/Caddyfile" /etc/caddy/Caddyfile
systemctl restart caddy

echo ""
echo "=== Setup complete ==="
