#!/bin/bash
set -e

echo "🚀 Starting VPS Setup for Madabank..."

# 1. Update System
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git unzip htop vi nano net-tools ufw fail2ban nginx

# 2. Install Docker
if ! command -v docker &> /dev/null; then
    echo "🐳 Installing Docker & Docker Compose..."
    sudo apt install -y ca-certificates curl gnupg
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg

    echo \
      "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
      sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    sudo apt update
    sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
else
    echo "✅ Docker already installed."
fi

# 3. Setup User Permissions
echo "👤 Configuring user permissions..."
sudo usermod -aG docker $USER || true

# 4. Security
echo "shield: Configuring Security..."
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
echo "y" | sudo ufw enable
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

echo "✅ Setup Complete! Please log out and log back in for Docker group changes to take effect."
