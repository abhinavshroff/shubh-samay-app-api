#!/usr/bin/env bash
# One-shot setup for Vultr Mumbai (Ubuntu 24.04)
# Run on a fresh VPS as root.
set -euo pipefail

echo "▶ Updating system…"
apt-get update -y && apt-get upgrade -y

echo "▶ Installing Docker…"
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

echo "▶ Cloning Shubh Samay backend…"
mkdir -p /opt/shubh-samay && cd /opt/shubh-samay
# Replace with your git repo URL
# git clone https://github.com/YOUR_ORG/shubh-samay.git .

echo "▶ Downloading Swiss Ephemeris data…"
mkdir -p ephemeris && cd ephemeris
curl -O https://www.astro.com/ftp/swisseph/ephe/sepl_18.se1
curl -O https://www.astro.com/ftp/swisseph/ephe/semo_18.se1
curl -O https://www.astro.com/ftp/swisseph/ephe/seas_18.se1
cd ..

echo "▶ Creating .env…"
cat > .env <<EOF
POSTGRES_PASSWORD=$(openssl rand -hex 16)
ADMIN_TOKEN=$(openssl rand -hex 24)
FCM_SERVER_KEY=YOUR_FCM_KEY_HERE
EOF

echo "▶ Building & starting…"
docker compose up -d --build

echo "▶ Setting up Caddy for HTTPS…"
apt-get install -y caddy
cat > /etc/caddy/Caddyfile <<EOF
api.shubhsamay.app {
  reverse_proxy localhost:8080
}
EOF
systemctl enable --now caddy
systemctl reload caddy

echo "▶ Setting up notification cron…"
(crontab -l 2>/dev/null; echo "* * * * * curl -s http://localhost:8080/v1/internal/dispatch-rahu") | crontab -
(crontab -l 2>/dev/null; echo "0 7 * * * curl -s http://localhost:8080/v1/internal/dispatch-morning") | crontab -

echo "✅ Done. Admin token saved in /opt/shubh-samay/.env"
echo "   API endpoint: https://api.shubhsamay.app/v1/health"
