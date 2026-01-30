# 1. Server Initialization Guide

**Goal:** Transform a fresh Ubuntu 24.04 VPS into a secured, Docker-ready server.

> [!TIP]
> **Automated Setup:** Run the provided helper script to automate Phases 1-4.
> ```bash
> # Run as root on your VPS
> curl -sSL https://raw.githubusercontent.com/darisadam/madabank-server/main/scripts/setup_vps.sh | bash
> ```

## Phase 1: Basic Provisioning

### 1. Update System
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git unzip htop vi nano net-tools ufw fail2ban
```

### 2. Create Admin User
Do not run services as `root`.
```bash
# Replace 'admin' with your preferred username
adduser admin
usermod -aG sudo admin
```

### 3. SSH Hardening
Edit `/etc/ssh/sshd_config`:
- `PermitRootLogin no`
- `PasswordAuthentication no` (Ensure you have copied your SSH keys first!)

Restart SSH: `sudo systemctl restart ssh`

### 4. Firewall (UFW)
```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 5. Install Docker
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker admin
```

**Verification:**
```bash
docker --version
docker compose version
```
