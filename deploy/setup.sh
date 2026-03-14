#!/bin/bash
# One-time Lightsail server setup (run as root or with sudo)
set -euo pipefail

echo "=== HexSlayer Lightsail Setup ==="

# System packages
apt-get update
apt-get install -y nginx

# Create app user and directories
useradd -r -s /bin/false hexslayer || true
mkdir -p /opt/hexslayer/frontend
mkdir -p /var/lib/hexslayer
chown hexslayer:hexslayer /var/lib/hexslayer

# Install systemd service
cp /tmp/hexslayer.service /etc/systemd/system/hexslayer.service
systemctl daemon-reload
systemctl enable hexslayer

# Install nginx config
rm -f /etc/nginx/sites-enabled/default
cp /tmp/hexslayer.nginx /etc/nginx/sites-available/hexslayer
ln -sf /etc/nginx/sites-available/hexslayer /etc/nginx/sites-enabled/hexslayer
nginx -t
systemctl enable nginx
systemctl restart nginx

echo "=== Setup complete ==="
echo "Now run 'make deploy' from your local machine to ship the build."
