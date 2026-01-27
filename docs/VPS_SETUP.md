# VPS Setup & Security Guide for Madabank Server
**Target OS:** Ubuntu 24.04 LTS

## 1. System Updates & Essential Packages
First, ensure your system is up to date and has the basic tools.

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git unzip htop vi nano net-tools
```

## 2. Docker Engine & Docker Compose
The application relies on Docker and Docker Compose (v2). We will install them from the official Ubuntu repositories.

```bash
# Install Docker and Compose plugin
sudo apt install -y docker.io docker-compose-v2

# Start and enable Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add your user to the docker group (avoids using sudo for docker commands)
sudo usermod -aG docker $USER
# NOTE: You must log out and log back in for this to take effect!
```

## 3. Web Server & SSL (Reverse Proxy)
We will use Nginx to reverse proxy traffic to your Docker container and handle SSL (HTTPS) via Let's Encrypt.

```bash
sudo apt install -y nginx certbot python3-certbot-nginx
```

## 4. Security Hardening Packages
### Firewall (UFW)
Ubuntu's default firewall tool.

```bash
sudo apt install -y ufw
```

### Intrusion Prevention (Fail2Ban)
Protects against brute-force attacks on SSH.

```bash
sudo apt install -y fail2ban
```

## 5. Database Tools (Optional/Utility)
Even though the database runs in Docker, having a local client is useful for debugging and backups.

```bash
sudo apt install -y postgresql-client redis-tools
```

---

## Security Configuration Steps

### 1. Configure Firewall (UFW)
By default, block everything and only allow essential ports.

```bash
# Set defaults
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (IMPORTANT: If you use a custom port, specify it here!)
sudo ufw allow ssh
# OR if using port 2222: sudo ufw allow 2222/tcp

# Allow Web Traffic
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable Firewall
sudo ufw enable
```

> **Warning:** Do NOT open ports 5432 (Postgres), 6379 (Redis), or 9090 (Prometheus) publicly. These should remain private within the Docker network or accessed via SSH Tunnel/VPN if needed.

### 2. Configure Fail2Ban
Fail2Ban is already effective with default settings for SSH on Ubuntu, but ensuring it's running is key.

```bash
sudo systemctl start fail2ban
sudo systemctl enable fail2ban
```

### 3. SSH Hardening (Check your `/etc/ssh/sshd_config`)
Ensure the following settings are set (edit with `sudo nano /etc/ssh/sshd_config`):
- `PermitRootLogin no` (You should log in as a regular user and use sudo)
- `PasswordAuthentication no` (Use SSH keys only)

Restart SSH after changes: `sudo systemctl restart ssh`

## Performance Tuning
### Add Swap Space
Docker containers for the stack (App + DB + Redis + Monitoring) can consume significant RAM. If your VPS has < 4GB RAM, a swap file is **highly recommended** to prevent crashes.

```bash
# Allocate 4GB swap
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

## Summary Checklist
- [ ] OS Updated
- [ ] Docker & Compose Installed
- [ ] Nginx Installed (Config will happen during deployment)
- [ ] UFW Enabled (Ports 22, 80, 443 Open only)
- [ ] Fail2Ban Running
- [ ] Swap File Created
