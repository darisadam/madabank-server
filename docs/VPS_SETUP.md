# VPS Setup & Security Guide for MadaBank Server
**Target OS:** Ubuntu 24.04 LTS

## 🐳 PHASE 1: SYSTEM PREPARATION
Ensure your system is up to date and has basic tools installed.

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git unzip htop vi nano net-tools
```

## 🛡️ PHASE 2: SECURITY BASICS

### 1. Install Security Tools
```bash
sudo apt install -y ufw fail2ban
```

### 2. Configure Firewall (UFW)
Block everything by default and only allow essential ports.

```bash
# Set defaults
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH (If you use a custom port like 2222, change 'ssh' to '2222/tcp')
sudo ufw allow ssh

# Allow Web Traffic
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable Firewall
sudo ufw enable
```

> **Warning:** Do NOT open ports 5432 (Postgres), 6379 (Redis), or 9090 (Prometheus) publicly.

### 3. SSH Hardening
Edit `/etc/ssh/sshd_config`:
- `PermitRootLogin no`
- `PasswordAuthentication no` (Use SSH Keys)
- Restart SSH: `sudo systemctl restart ssh`

## 🌐 PHASE 3: WEB SERVER (REVERSE PROXY)
We use Nginx to reverse proxy traffic to Docker containers and handle SSL.

```bash
sudo apt install -y nginx certbot python3-certbot-nginx
```

## 🐳 PHASE 4: INSTALL DOCKER & DOCKER COMPOSE

### Step 14: Install Docker
If you have an older version of Docker installed, it's recommended to uninstall it first:
```bash
sudo apt-get remove docker docker-engine docker.io containerd runc
```

**Install Docker (Official Method):**

Download and run the install script:
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
```

**This process will:**
- Detect your OS (Ubuntu)
- Add the Docker repository
- Install Docker Engine
- Setup Docker service

**Cleanup script:**
```bash
rm get-docker.sh
```

### Step 15: Add User to Docker Group
Allow the `admin` user to run Docker commands without sudo.

```bash
sudo usermod -aG docker admin
```

**Action Required:** Logout and login again for group membership to apply.

**Verification:**
```bash
docker --version
# Should show: Docker version 24.x.x or later

docker run hello-world
# Should show: "Hello from Docker!"
```

### Step 16: Install Docker Compose
Docker Compose V2 is typically included with modern Docker installations.

**Verify:**
```bash
docker compose version
# Should show: Docker Compose version v2.x.x
```

**If missing (Manual Install):**
```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.5/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Step 17: Configure Docker (Production Ready)
Setup Docker daemon logging to prevent disk filling.

Edit config: `sudo nano /etc/docker/daemon.json`
```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2"
}
```

Restart Docker:
```bash
sudo systemctl restart docker
sudo systemctl status docker
```

---

## 🚀 PHASE 5: SETUP APPLICATION STRUCTURE

### Step 18: Create Application User
Create a dedicated user `bankingapp` with no login access for security.

```bash
# Create user with no login shell
sudo useradd -r -m -d /opt/bankingapp -s /usr/sbin/nologin bankingapp
```

**Verify:**
```bash
id bankingapp
# Should show: uid=999(bankingapp) ...
```

### Step 19: Setup Directory Structure
Create the folder hierarchy for the application.

```bash
sudo mkdir -p /opt/bankingapp/{app,logs,backups,scripts,nginx,prometheus,grafana,loki}
sudo mkdir -p /opt/bankingapp/backups/{postgres,app-data}
sudo mkdir -p /opt/bankingapp/nginx/conf.d

# Verify
ls -la /opt/bankingapp/
```

### Step 20: Set Ownership & Permissions
Set `bankingapp` as the owner.

```bash
sudo chown -R bankingapp:bankingapp /opt/bankingapp
```

Set secure permissions:
```bash
# Base directory (750)
sudo chmod 750 /opt/bankingapp

# Logs (770 - writable)
sudo chmod 770 /opt/bankingapp/logs

# Backups (750 - protected)
sudo chmod 750 /opt/bankingapp/backups

# Scripts (750)
sudo chmod 750 /opt/bankingapp/scripts
```

### Step 21: Add Admin to bankingapp Group
Allow `admin` to manage application files.

```bash
sudo usermod -aG bankingapp admin
```

**Action Required:** Logout and login again.

### Step 22: Create Environment Template
Create `.env.example` as a template for secrets.

`sudo nano /opt/bankingapp/.env.example`

```bash
# ==========================================
# BANKING APP - ENVIRONMENT VARIABLES
# ==========================================

# Application
NODE_ENV=production
PORT=3000
APP_NAME=Banking App
LOG_LEVEL=info

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=bankingapp_prod
DB_USER=bankingapp_user
DB_PASSWORD=placeholder_db_password

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=placeholder_redis_password

# JWT & Secrets
JWT_SECRET=placeholder_jwt_secret_min_64_chars
JWT_EXPIRY=24h
REFRESH_TOKEN_SECRET=placeholder_refresh_secret
ENCRYPTION_KEY=placeholder_encryption_key_32_chars

# API Rate Limiting
RATE_LIMIT_WINDOW=15m
RATE_LIMIT_MAX_REQUESTS=100

# Email
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=placeholder_smtp_password
EMAIL_FROM=noreply@yourdomain.com

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_PASSWORD=placeholder_grafana_password

# Backup
BACKUP_RETENTION_DAYS=30
```

Set permissions:
```bash
chmod 640 /opt/bankingapp/.env.example
```

### Step 23: Setup Swap Space
Optimize RAM usage with 2GB swap (or more if needed).

```bash
# Check existing
sudo swapon --show

# Create 2GB swap file
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Optimize Swappiness (Set to 10 for servers)
sudo sysctl vm.swappiness=10
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
```

### Step 24: System Limits & Optimization
Increase file descriptors for high-load Docker usage.

**1. Edit Limits:** `sudo nano /etc/security/limits.conf`
Add to end:
```conf
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
```

**2. Optimize Kernel:** `sudo nano /etc/sysctl.conf`
Add to end:
```conf
# Network optimization
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.ip_local_port_range = 1024 65535

# File system
fs.file-max = 2097152

# Docker optimization (Required for some DBs/ES)
vm.max_map_count = 262144
```

**Apply changes:**
```bash
sudo sysctl -p
```

### Step 25: Create Docker Networks
Isolate containers by function.

```bash
docker network create frontend
docker network create backend
docker network create monitoring
```

### Step 26: Create Helper Scripts

**1. System Status Script:** `sudo nano /opt/bankingapp/scripts/system-status.sh`

```bash
#!/bin/bash
echo "========================================="
echo "  BANKING APP - SYSTEM STATUS"
echo "========================================="
echo ""
echo "Memory Usage:"
free -h
echo ""
echo "Disk Usage:"
df -h /
echo ""
echo "Swap Usage:"
swapon --show
echo ""
echo "Docker Containers:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "Docker Networks:"
docker network ls
echo ""
echo "Firewall Status:"
sudo ufw status
echo ""
echo "Failed Login Attempts (last 10):"
sudo tail -10 /var/log/auth.log | grep "Failed password"
echo ""
echo "========================================="
```

**2. Docker Status Script:** `sudo nano /opt/bankingapp/scripts/docker-status.sh`

```bash
#!/bin/bash
echo "Docker Containers Status:"
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Image}}"
echo ""
echo "Docker Images:"
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
echo ""
echo "Docker Volumes:"
docker volume ls
echo ""
echo "Docker Disk Usage:"
docker system df
```

**Make executable:**
```bash
sudo chmod +x /opt/bankingapp/scripts/system-status.sh
sudo chmod +x /opt/bankingapp/scripts/docker-status.sh
```

## ✅ CHECKPOINT
At the end of Phase 5, you should have:
- Docker & Compose installed and configured
- `bankingapp` user and directory structure created
- `admin` user added to necessary groups
- System optimized (swap, limits)
- Networks created

---

## 🚀 PHASE 6: DEPLOY SERVICES WITH DOCKER COMPOSE

**Goal:**
- Create `docker-compose.yml` for production
- Configure Nginx and PostgreSQL
- Setup `.env` file with generated secrets
- Pull and start all services

**Estimated time:** 30-40 minutes

### Step 27: Generate Secure Secrets
Before creating configuration files, generate strong secrets. Run these commands on your local machine or the VPS and **save the outputs securely** (e.g., in a password manager).

**Generate Secrets:**
```bash
# Database Password
openssl rand -base64 32

# Redis Password
openssl rand -base64 32

# JWT Secret (64 chars)
openssl rand -base64 64

# Refresh Token Secret
openssl rand -base64 64

# Encryption Key
openssl rand -base64 32

# Grafana Password
openssl rand -base64 16
```

### Step 28: Create Production .env File
Create the environment file.

`sudo nano /opt/bankingapp/.env`

Paste the template below and **replace** all `PASTE_XXX_HERE` placeholders with your generated secrets.

```bash
# ==========================================
# MADABANK APP - PRODUCTION ENVIRONMENT
# ==========================================

# Application
NODE_ENV=production
PORT=3000
APP_NAME=MadaBank App
LOG_LEVEL=info

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=bankingapp_prod
DB_USER=bankingapp_user
DB_PASSWORD=PASTE_DB_PASSWORD_HERE

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=PASTE_REDIS_PASSWORD_HERE

# JWT & Secrets
JWT_SECRET=PASTE_JWT_SECRET_HERE
JWT_EXPIRY=24h
REFRESH_TOKEN_SECRET=PASTE_REFRESH_TOKEN_SECRET_HERE
ENCRYPTION_KEY=PASTE_ENCRYPTION_KEY_HERE

# API Rate Limiting
RATE_LIMIT_WINDOW=15m
RATE_LIMIT_MAX_REQUESTS=100

# Email (SMTP)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=YOUR_SENDGRID_API_KEY_WHEN_READY
EMAIL_FROM=noreply@yourdomain.com

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_PASSWORD=PASTE_GRAFANA_PASSWORD_HERE

# Backup
BACKUP_RETENTION_DAYS=30
```

**Secure the file:**
```bash
sudo chmod 600 /opt/bankingapp/.env
sudo chown bankingapp:bankingapp /opt/bankingapp/.env

# Verify (Should be -rw------- 1 bankingapp bankingapp)
ls -la /opt/bankingapp/.env
```

### Step 29: Create docker-compose.yml (Part 1 - Infrastructure)
Create the Compose file.

`sudo nano /opt/bankingapp/docker-compose.yml`

Paste the infrastructure configuration:

```yaml
version: '3.8'

services:
  # ============================================
  # PostgreSQL Database
  # ============================================
  postgres:
    image: postgres:15-alpine
    container_name: postgres
    restart: always
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./backups/postgres:/backups
    networks:
      - backend
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 2560M
          cpus: '0.5'
        reservations:
          memory: 2G
          cpus: '0.3'
    shm_size: 256mb

  # ============================================
  # Redis Cache
  # ============================================
  redis:
    image: redis:7-alpine
    container_name: redis
    restart: always
    command: >
      redis-server
      --maxmemory 512mb
      --maxmemory-policy allkeys-lru
      --appendonly yes
      --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    networks:
      - backend
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.1'

# ============================================
# Networks
# ============================================
networks:
  frontend:
    external: true
  backend:
    external: true
  monitoring:
    external: true

# ============================================
# Volumes
# ============================================
volumes:
  postgres-data:
    driver: local
  redis-data:
    driver: local
```

### Step 30: Test Database Services
Verify Postgres and Redis before proceeding.

```bash
cd /opt/bankingapp
docker compose pull postgres redis
docker compose up -d postgres redis

# Check status (Should be 'healthy')
docker compose ps
```

**Test Connections:**
```bash
# Test Postgres
docker exec -it postgres psql -U bankingapp_user -d bankingapp_prod -c "SELECT version();"

# Test Redis
docker exec -it redis redis-cli -a $(grep REDIS_PASSWORD /opt/bankingapp/.env | cut -d '=' -f2) PING
# Should return PONG
```

### Step 31: Setup Nginx Configuration
Create main config: `sudo nano /opt/bankingapp/nginx/nginx.conf`

```nginx
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 4096;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main;

    # Performance & Security
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 10M;

    # Gzip
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=3r/m;
    limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

    include /etc/nginx/conf.d/*.conf;
}
```

Create site config: `sudo nano /opt/bankingapp/nginx/conf.d/madabank.conf`

```nginx
# Upstream (Uncomment after app deployment)
# upstream backend {
#     least_conn;
#     server app:3000 max_fails=3 fail_timeout=30s;
#     keepalive 32;
# }

# HTTP Server (Test config)
server {
    listen 80;
    listen [::]:80;
    server_name _;

    location /health {
        access_log off;
        return 200 "Nginx OK\n";
        add_header Content-Type text/plain;
    }

    location / {
        return 200 "MadaBank Server Ready\n";
        add_header Content-Type text/plain;
    }
}
```

### Step 32: Add Nginx to Docker Compose
Edit `sudo nano /opt/bankingapp/docker-compose.yml`. Add the `nginx` service (after redis):

```yaml
  # ============================================
  # Nginx Reverse Proxy
  # ============================================
  nginx:
    image: nginx:alpine
    container_name: nginx
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - nginx-logs:/var/log/nginx
    networks:
      - frontend
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.2'
```

Add to `volumes` section at bottom:
```yaml
  nginx-logs:
    driver: local
```

**Start Nginx:**
```bash
docker compose up -d nginx
curl http://localhost/health
# Should return: Nginx OK
```

### Step 33: Add Monitoring Stack (Prometheus)
Create Config: `sudo nano /opt/bankingapp/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
```

Add to `docker-compose.yml`:

```yaml
  # ============================================
  # Prometheus (Metrics)
  # ============================================
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    restart: always
    user: "65534:65534"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    ports:
      - "127.0.0.1:9090:9090"
    networks:
      - monitoring
      - backend
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.2'

  # ============================================
  # Node Exporter (System Metrics)
  # ============================================
  node-exporter:
    image: prom/node-exporter:latest
    container_name: node-exporter
    restart: always
    command:
      - '--path.procfs=/host/proc'
      - '--path.sysfs=/host/sys'
      - '--path.rootfs=/rootfs'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    ports:
      - "127.0.0.1:9100:9100"
    networks:
      - monitoring
    deploy:
      resources:
        limits:
          memory: 64M
          cpus: '0.05'

  # ============================================
  # PostgreSQL Exporter (DB Metrics)
  # ============================================
  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:latest
    container_name: postgres-exporter
    restart: always
    environment:
      DATA_SOURCE_NAME: "postgresql://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable"
    ports:
      - "127.0.0.1:9187:9187"
    depends_on:
      - postgres
    networks:
      - monitoring
      - backend
    deploy:
      resources:
        limits:
          memory: 64M
          cpus: '0.05'
```

Add `prometheus-data` to volumes. Start stack: `docker compose up -d prometheus node-exporter postgres-exporter`.

---

## 📊 PHASE 7: MONITORING & VISUALIZATION

### Step 34: Add Grafana
Add to `docker-compose.yml`:

```yaml
  # ============================================
  # Grafana (Visualization)
  # ============================================
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: always
    user: "472:472"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_SERVER_ROOT_URL=http://localhost:3001
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning:ro
    ports:
      - "127.0.0.1:3001:3000"
    depends_on:
      - prometheus
    networks:
      - monitoring
    deploy:
      resources:
        limits:
          memory: 300M
          cpus: '0.2'
```

Add `grafana-data` to volumes.

**Setup Data Source:**
`sudo mkdir -p /opt/bankingapp/grafana/provisioning/{datasources,dashboards}`
`sudo nano /opt/bankingapp/grafana/provisioning/datasources/prometheus.yml`

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

Start Grafana: `docker compose up -d grafana`.

### Step 35: Add Logging (Loki + Promtail)
**Loki Config:** `sudo nano /opt/bankingapp/loki/loki-config.yaml`
_(Content omitted for brevity - use standard BoltDB-shipper config)_

**Promtail Config:** `sudo nano /opt/bankingapp/promtail/promtail-config.yaml`
_(Content omitted for brevity - scrape /var/lib/docker/containers)_

Add both to `docker-compose.yml` and `loki-data` to volumes. Start services.

### Step 36: Automated Backups
**1. Database Backup Script:** `sudo nano /opt/bankingapp/scripts/backup-database.sh`
```bash
#!/bin/bash
set -e
BACKUP_DIR="/opt/bankingapp/backups/postgres"
DATE=$(date +%Y%m%d-%H%M%S)
# ... [Full script content available in previous steps] ...
```
Make executable: `chmod +x /opt/bankingapp/scripts/backup-database.sh`.

**2. Configure Cron:**
`crontab -e`
```bash
# Database backup - daily at 2:00 AM
0 2 * * * /opt/bankingapp/scripts/backup-database.sh >> /opt/bankingapp/logs/backup.log 2>&1
```

---

## 🔒 PHASE 8: SSL/HTTPS & DOMAIN SETUP

### Step 42: Configure DNS
1.  **Login to your Domain Registrar** (e.g., WordPress.com, Namecheap).
2.  **Add 'A' Records:**
    *   **Host:** `@` (Root) -> **Value:** `YOUR_VPS_IP`
    *   **Host:** `api` -> **Value:** `YOUR_VPS_IP`
    *   **Host:** `monitoring` -> **Value:** `YOUR_VPS_IP`
3.  **Wait for Propagation:** This may take 15 minutes to 24 hours.

### Step 43: Install SSL with Certbot
Once DNS is propagating, obtain certificates.

```bash
# Stop Nginx temporarily to allow Certbot standalone verification OR use --nginx plugin
sudo certbot --nginx -d madabank.art -d api.madabank.art -d monitoring.madabank.art
```

### Step 44: Finalize Nginx Config
Update `madabank.conf` to specific blocks for `api`, `monitoring`, and `root`.

**Example 'api' block:**
```nginx
server {
    listen 443 ssl http2;
    server_name api.madabank.art;

    ssl_certificate /etc/letsencrypt/live/madabank.art/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/madabank.art/privkey.pem;

    location / {
        proxy_pass http://backend:8080; # Ensure 'app' service is running on port 8080
    }
}
```

Restart Nginx: `docker compose restart nginx`.

## ✅ Final Verification
1.  **API:** `https://api.madabank.art/health` -> 200 OK
2.  **Monitoring:** `https://monitoring.madabank.art` -> Grafana Login
3.  **Database:** Clean backups in `/opt/bankingapp/backups`

---

## 🚀 PHASE 9: CI/CD DEPLOYMENT WITH GITHUB ACTIONS

This section explains how to set up automated CI/CD deployment for MadaBank Server using GitHub Actions.

### Overview
The CI/CD pipeline consists of two workflows:
- **CI Pipeline (`ci.yml`):** Runs on every push/PR to `main` or `develop` branches. Performs linting, unit tests, integration tests, security scanning, and Docker build validation.
- **CD Pipeline:** Triggers after successful CI on `main` branch to deploy to VPS automatically.

### Step 45: Prerequisites
Before setting up CI/CD, ensure:
1. ✅ VPS is configured following Phases 1-8
2. ✅ Docker and Docker Compose installed on VPS
3. ✅ GitHub repository with MadaBank server code
4. ✅ Domain configured with SSL (Phase 8)

### Step 46: Generate SSH Deploy Key
Create a dedicated SSH key pair for GitHub Actions to access your VPS.

**On your local machine:**
```bash
# Generate a new SSH key pair (Ed25519 recommended)
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/github_actions_deploy -N ""

# Display the public key (add to VPS)
cat ~/.ssh/github_actions_deploy.pub

# Display the private key (add to GitHub Secrets)
cat ~/.ssh/github_actions_deploy
```

**On VPS - Add public key to authorized_keys:**
```bash
# Switch to admin user (or the user that runs deployments)
sudo su - admin

# Add the public key
echo "PASTE_PUBLIC_KEY_HERE" >> ~/.ssh/authorized_keys

# Verify permissions
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

### Step 47: Configure GitHub Repository Secrets
Navigate to your GitHub repository → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**.

Add the following secrets:

| Secret Name | Description | Example Value |
|-------------|-------------|---------------|
| `VPS_HOST` | VPS IP address or domain | `123.45.67.89` or `madabank.art` |
| `VPS_USER` | SSH username for deployment | `admin` |
| `VPS_SSH_KEY` | Private SSH key (from Step 46) | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `VPS_SSH_PORT` | SSH port (default 22) | `22` |
| `DOCKER_REGISTRY` | Container registry URL | `ghcr.io/yourusername` |
| `DOCKER_USERNAME` | Registry username | `yourusername` |
| `DOCKER_PASSWORD` | Registry token/password | `ghp_xxxxx` (GitHub PAT) |

> **Tip:** For `DOCKER_PASSWORD`, use a GitHub Personal Access Token (PAT) with `write:packages` scope.

### Step 48: Create Deployment Script on VPS
Create the deployment script that GitHub Actions will execute.

**Create script:** `sudo nano /opt/bankingapp/scripts/deploy.sh`
```bash
#!/bin/bash
set -e

# Configuration
APP_DIR="/opt/bankingapp"
COMPOSE_FILE="$APP_DIR/docker-compose.yml"
REGISTRY="${DOCKER_REGISTRY:-ghcr.io/yourusername}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

echo "========================================="
echo "  MADABANK DEPLOYMENT - $(date)"
echo "========================================="

# Step 1: Login to Docker Registry
echo "[1/5] Logging into Docker registry..."
echo "$DOCKER_PASSWORD" | docker login ghcr.io -u "$DOCKER_USERNAME" --password-stdin

# Step 2: Pull latest images
echo "[2/5] Pulling latest images..."
docker compose -f "$COMPOSE_FILE" pull app

# Step 3: Stop existing containers
echo "[3/5] Stopping existing containers..."
docker compose -f "$COMPOSE_FILE" stop app

# Step 4: Start updated containers
echo "[4/5] Starting updated containers..."
docker compose -f "$COMPOSE_FILE" up -d app

# Step 5: Cleanup old images
echo "[5/5] Cleaning up old images..."
docker image prune -f

# Health check
echo ""
echo "Waiting for health check..."
sleep 10
curl -s http://localhost:8080/health || echo "Health check pending..."

echo ""
echo "✅ Deployment completed successfully!"
echo "========================================="
```

**Make executable:**
```bash
sudo chmod +x /opt/bankingapp/scripts/deploy.sh
sudo chown admin:admin /opt/bankingapp/scripts/deploy.sh
```

### Step 49: Add App Service to docker-compose.yml
Add the MadaBank API service to your `docker-compose.yml`:

```yaml
  # ============================================
  # MadaBank API Application
  # ============================================
  app:
    image: ${DOCKER_REGISTRY}/madabank-api:${IMAGE_TAG:-latest}
    container_name: madabank-api
    restart: always
    env_file:
      - .env
    environment:
      - DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - ENV=production
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - frontend
      - backend
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
```

### Step 50: Create CD Workflow
Create or update `.github/workflows/cd.yml` for deployment:

```yaml
name: CD Pipeline

on:
  push:
    branches: [ main ]
  workflow_dispatch:  # Allow manual trigger

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository_owner }}/madabank-api

jobs:
  # Build and push Docker image
  build-and-push:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    outputs:
      image_tag: ${{ steps.meta.outputs.tags }}
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Build Binaries
        env:
          GOOS: linux
          GOARCH: amd64
          CGO_ENABLED: 0
        run: |
          go build -ldflags "-s -w" -o bin/api-linux-amd64 cmd/api/main.go
          go build -ldflags "-s -w" -o bin/migrate-linux-amd64 cmd/migrate/main.go

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=sha,prefix=
            type=raw,value=latest

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./docker/Dockerfile.fast
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          platforms: linux/amd64

  # Deploy to VPS
  deploy:
    name: Deploy to VPS
    runs-on: ubuntu-latest
    needs: build-and-push
    environment: production
    
    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          port: ${{ secrets.VPS_SSH_PORT }}
          script: |
            export DOCKER_USERNAME="${{ github.actor }}"
            export DOCKER_PASSWORD="${{ secrets.GITHUB_TOKEN }}"
            export DOCKER_REGISTRY="${{ env.REGISTRY }}"
            export IMAGE_TAG="${{ github.sha }}"
            /opt/bankingapp/scripts/deploy.sh

      - name: Verify Deployment
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          port: ${{ secrets.VPS_SSH_PORT }}
          script: |
            echo "Checking container status..."
            docker ps | grep madabank-api
            echo ""
            echo "Checking API health..."
            curl -s http://localhost:8080/health
```

### Step 51: Update Nginx for App Proxy
Update `/opt/bankingapp/nginx/conf.d/madabank.conf` to proxy to the app:

```nginx
# Upstream backend
upstream backend {
    least_conn;
    server app:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

# HTTPS Server for API
server {
    listen 443 ssl http2;
    server_name api.madabank.art;

    ssl_certificate /etc/letsencrypt/live/madabank.art/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/madabank.art/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Rate limiting for API
    limit_req zone=api_limit burst=20 nodelay;
    limit_conn conn_limit 10;

    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 90s;
        proxy_connect_timeout 90s;
    }

    # Health check endpoint (no rate limit)
    location /health {
        limit_req off;
        proxy_pass http://backend;
    }
}
```

Restart Nginx after update:
```bash
docker compose restart nginx
```

### Step 52: Test the CI/CD Pipeline
1. **Push to main branch:**
   ```bash
   git add .
   git commit -m "chore: trigger deployment"
   git push origin main
   ```

2. **Monitor GitHub Actions:**
   - Go to repository → **Actions** tab
   - Watch CI Pipeline run (lint, test, build)
   - Watch CD Pipeline run (build image, deploy)

3. **Verify on VPS:**
   ```bash
   # Check container status
   docker ps | grep madabank

   # Check logs
   docker logs madabank-api --tail 50

   # Test API
   curl https://api.madabank.art/health
   ```

### Step 53: Rollback Procedure
If deployment fails, rollback to previous version:

```bash
# On VPS
cd /opt/bankingapp

# List available image tags
docker images | grep madabank-api

# Update compose to use previous tag
export IMAGE_TAG="previous_commit_sha"

# Restart with previous version
docker compose up -d app

# Verify
curl http://localhost:8080/health
```

### CI/CD Workflow Diagram
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Developer     │────▶│   GitHub Repo   │────▶│  GitHub Actions │
│   Push Code     │     │   (main branch) │     │   (CI Pipeline) │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                         │
                        ┌────────────────────────────────┘
                        ▼
              ┌─────────────────┐     ┌─────────────────┐
              │  Build Docker   │────▶│  Push to GHCR   │
              │     Image       │     │  (Container     │
              └─────────────────┘     │   Registry)     │
                                      └────────┬────────┘
                                               │
                                               ▼
                                      ┌─────────────────┐
                                      │  CD Pipeline    │
                                      │  (SSH to VPS)   │
                                      └────────┬────────┘
                                               │
                                               ▼
                                      ┌─────────────────┐
                                      │  VPS: Pull &    │
                                      │  Deploy Image   │
                                      └────────┬────────┘
                                               │
                                               ▼
                                      ┌─────────────────┐
                                      │  ✅ Live at     │
                                      │  api.madabank   │
                                      └─────────────────┘
```

## ✅ CI/CD Checkpoint (GitHub Actions)
After completing Phase 9, you should have:
- ✅ SSH deploy key configured
- ✅ GitHub repository secrets set up
- ✅ Deployment script on VPS
- ✅ CD workflow configured
- ✅ Nginx proxying to app container
- ✅ Successful automated deployment on push to `main`

---

## 🔧 PHASE 10: JENKINS CI/CD ON VPS WITH GITHUB WEBHOOK

This section explains how to set up Jenkins on your VPS for CI/CD with GitHub webhook integration as an alternative to GitHub Actions.

### Overview
With Jenkins on VPS:
- Jenkins runs directly on your VPS
- GitHub webhook triggers builds on push/PR events
- Jenkins builds Docker images locally and deploys immediately
- No external CI/CD service dependency

### Step 54: Install Java (Jenkins Requirement)
Jenkins requires Java 17 or 21.

```bash
# Update packages
sudo apt update

# Install OpenJDK 17
sudo apt install -y openjdk-17-jdk

# Verify installation
java -version
# Should show: openjdk version "17.x.x"
```

### Step 55: Install Jenkins
Add Jenkins repository and install.

```bash
# Add Jenkins GPG key
sudo wget -O /usr/share/keyrings/jenkins-keyring.asc \
  https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key

# Add Jenkins repository
echo "deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc]" \
  "https://pkg.jenkins.io/debian-stable binary/" | sudo tee \
  /etc/apt/sources.list.d/jenkins.list > /dev/null

# Install Jenkins
sudo apt update
sudo apt install -y jenkins

# Start and enable Jenkins
sudo systemctl start jenkins
sudo systemctl enable jenkins
sudo systemctl status jenkins
```

### Step 56: Configure Firewall for Jenkins
Allow Jenkins port (8080) for initial setup, then restrict later.

```bash
# Temporarily allow Jenkins port (for initial setup only)
sudo ufw allow 8080/tcp

# Verify
sudo ufw status
```

> **Security Note:** After setup, we'll configure Nginx reverse proxy and remove direct port access.

### Step 57: Initial Jenkins Setup
1. **Get initial admin password:**
   ```bash
   sudo cat /var/lib/jenkins/secrets/initialAdminPassword
   ```

2. **Access Jenkins:** Open `http://YOUR_VPS_IP:8080` in browser

3. **Complete setup wizard:**
   - Paste the initial admin password
   - Select "Install suggested plugins"
   - Create admin user (save credentials securely)
   - Set Jenkins URL: `http://YOUR_VPS_IP:8080` (will update to HTTPS later)

### Step 58: Install Required Jenkins Plugins
Navigate to **Manage Jenkins** → **Plugins** → **Available plugins**.

Install these plugins:
- **Git** - Git integration
- **GitHub** - GitHub integration
- **GitHub Integration** - Webhook support
- **Pipeline** - Pipeline as code
- **Docker Pipeline** - Docker support in pipelines
- **Docker** - Docker build and publish
- **Credentials Binding** - Secure credential handling
- **Blue Ocean** - Modern UI (optional but recommended)

Click **Install** and restart Jenkins when prompted.

### Step 59: Add Jenkins User to Docker Group
Allow Jenkins to run Docker commands.

```bash
# Add jenkins user to docker group
sudo usermod -aG docker jenkins

# Restart Jenkins to apply
sudo systemctl restart jenkins

# Verify
sudo -u jenkins docker ps
```

### Step 60: Configure Jenkins Credentials
Navigate to **Manage Jenkins** → **Credentials** → **System** → **Global credentials** → **Add Credentials**.

Add the following credentials:

**1. GitHub Personal Access Token:**
- Kind: `Secret text`
- Scope: `Global`
- Secret: `YOUR_GITHUB_PAT` (with `repo` and `admin:repo_hook` scopes)
- ID: `github-token`
- Description: `GitHub Personal Access Token`

**2. Docker Registry Credentials (if using external registry):**
- Kind: `Username with password`
- Scope: `Global`
- Username: `YOUR_USERNAME`
- Password: `YOUR_TOKEN`
- ID: `docker-registry`
- Description: `Docker Registry Credentials`

### Step 61: Configure GitHub Webhook
**On GitHub Repository:**

1. Go to **Settings** → **Webhooks** → **Add webhook**
2. Configure:
   - **Payload URL:** `http://YOUR_VPS_IP:8080/github-webhook/`
   - **Content type:** `application/json`
   - **Secret:** Generate a secure secret and save it
     ```bash
     openssl rand -hex 32
     ```
   - **Events:** Select "Just the push event" or "Let me select individual events" (Push, Pull Request)
3. Click **Add webhook**

**On Jenkins:**

1. Go to **Manage Jenkins** → **System**
2. Find **GitHub** section
3. Add GitHub Server:
   - Name: `GitHub`
   - API URL: `https://api.github.com`
   - Credentials: Select `github-token`
4. Check "Manage hooks"
5. Click **Save**

### Step 62: Create Jenkins Pipeline Job
1. Click **New Item**
2. Enter name: `madabank-server`
3. Select **Pipeline**
4. Click **OK**

**Configure the job:**

**General:**
- ✅ GitHub project
- Project url: `https://github.com/YOUR_USERNAME/madabank-server`

**Build Triggers:**
- ✅ GitHub hook trigger for GITScm polling

**Pipeline:**
- Definition: `Pipeline script from SCM`
- SCM: `Git`
- Repository URL: `https://github.com/YOUR_USERNAME/madabank-server.git`
- Credentials: Select `github-token`
- Branch Specifier: `*/main` (or `*/develop` for dev builds)
- Script Path: `Jenkinsfile`

Click **Save**.

### Step 63: Create Jenkinsfile
Create `Jenkinsfile` in your repository root:

```groovy
pipeline {
    agent any
    
    environment {
        APP_NAME = 'madabank-api'
        DOCKER_IMAGE = 'madabank-api'
        DEPLOY_DIR = '/opt/bankingapp'
        GO_VERSION = '1.24'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
                sh 'git log -1 --pretty=format:"%h - %s (%an)"'
            }
        }
        
        stage('Setup Go') {
            steps {
                sh '''
                    # Download and install Go if not present
                    if ! command -v go &> /dev/null || [ "$(go version | grep -oP '\\d+\\.\\d+')" != "${GO_VERSION}" ]; then
                        wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
                        sudo rm -rf /usr/local/go
                        sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
                        rm go${GO_VERSION}.linux-amd64.tar.gz
                    fi
                    export PATH=$PATH:/usr/local/go/bin
                    go version
                '''
            }
        }
        
        stage('Install Dependencies') {
            steps {
                sh '''
                    export PATH=$PATH:/usr/local/go/bin
                    go mod download
                    go mod verify
                '''
            }
        }
        
        stage('Lint') {
            steps {
                sh '''
                    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
                    # Install golangci-lint if not present
                    if ! command -v golangci-lint &> /dev/null; then
                        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $HOME/go/bin v1.64.5
                    fi
                    golangci-lint run --timeout=5m
                '''
            }
        }
        
        stage('Test') {
            steps {
                sh '''
                    export PATH=$PATH:/usr/local/go/bin
                    go test -v -race -coverprofile=coverage.out ./...
                '''
            }
            post {
                always {
                    sh 'go tool cover -html=coverage.out -o coverage.html || true'
                    archiveArtifacts artifacts: 'coverage.html', allowEmptyArchive: true
                }
            }
        }
        
        stage('Build Binary') {
            steps {
                sh '''
                    export PATH=$PATH:/usr/local/go/bin
                    export CGO_ENABLED=0
                    export GOOS=linux
                    export GOARCH=amd64
                    
                    mkdir -p bin
                    go build -ldflags "-s -w" -o bin/api-linux-amd64 cmd/api/main.go
                    go build -ldflags "-s -w" -o bin/migrate-linux-amd64 cmd/migrate/main.go
                    
                    ls -la bin/
                '''
            }
        }
        
        stage('Build Docker Image') {
            steps {
                sh '''
                    docker build -f docker/Dockerfile.fast -t ${DOCKER_IMAGE}:${BUILD_NUMBER} .
                    docker tag ${DOCKER_IMAGE}:${BUILD_NUMBER} ${DOCKER_IMAGE}:latest
                '''
            }
        }
        
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                sh '''
                    echo "Deploying to production..."
                    
                    # Stop existing container
                    docker compose -f ${DEPLOY_DIR}/docker-compose.yml stop app || true
                    
                    # Update image tag in compose
                    export IMAGE_TAG=${BUILD_NUMBER}
                    
                    # Start new container
                    docker compose -f ${DEPLOY_DIR}/docker-compose.yml up -d app
                    
                    # Wait for health check
                    sleep 15
                    
                    # Verify deployment
                    curl -f http://localhost:8080/health || exit 1
                    
                    echo "✅ Deployment successful!"
                '''
            }
        }
        
        stage('Cleanup') {
            steps {
                sh '''
                    # Remove old images (keep last 3)
                    docker images ${DOCKER_IMAGE} --format "{{.ID}} {{.Tag}}" | \
                        sort -t. -k1 -n | head -n -3 | \
                        awk '{print $1}' | xargs -r docker rmi || true
                    
                    # Prune dangling images
                    docker image prune -f
                '''
            }
        }
    }
    
    post {
        success {
            echo '✅ Pipeline completed successfully!'
        }
        failure {
            echo '❌ Pipeline failed!'
            // Optional: Add notification (email, Slack, etc.)
        }
        always {
            cleanWs()
        }
    }
}
```

### Step 64: Update docker-compose.yml for Jenkins
Update the app service to use local images:

```yaml
  # ============================================
  # MadaBank API Application (Jenkins Build)
  # ============================================
  app:
    image: madabank-api:${IMAGE_TAG:-latest}
    container_name: madabank-api
    restart: always
    env_file:
      - .env
    environment:
      - DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - ENV=production
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - frontend
      - backend
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Step 65: Configure Nginx Reverse Proxy for Jenkins
Secure Jenkins behind Nginx with SSL.

**Create Jenkins Nginx config:** `sudo nano /opt/bankingapp/nginx/conf.d/jenkins.conf`

```nginx
# Jenkins Reverse Proxy
server {
    listen 443 ssl http2;
    server_name jenkins.madabank.art;

    ssl_certificate /etc/letsencrypt/live/madabank.art/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/madabank.art/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Required for Jenkins websocket agents
        proxy_set_header Connection "upgrade";
        proxy_set_header Upgrade $http_upgrade;
        
        proxy_read_timeout 90s;
        proxy_buffering off;
    }
}

# HTTP redirect
server {
    listen 80;
    server_name jenkins.madabank.art;
    return 301 https://$server_name$request_uri;
}
```

**Get SSL certificate for Jenkins subdomain:**
```bash
sudo certbot --nginx -d jenkins.madabank.art
```

**Remove direct port access:**
```bash
sudo ufw delete allow 8080/tcp
sudo ufw status
```

**Restart Nginx:**
```bash
docker compose restart nginx
```

**Update Jenkins URL:**
1. Go to **Manage Jenkins** → **System**
2. Update Jenkins URL to `https://jenkins.madabank.art/`
3. Save

**Update GitHub Webhook:**
Update webhook URL to `https://jenkins.madabank.art/github-webhook/`

### Step 66: Test the Jenkins Pipeline
1. **Push to repository:**
   ```bash
   git add .
   git commit -m "chore: test jenkins pipeline"
   git push origin main
   ```

2. **Monitor Jenkins:**
   - Open `https://jenkins.madabank.art`
   - Check `madabank-server` job
   - Watch build progress in Blue Ocean or Classic UI

3. **Verify deployment:**
   ```bash
   # On VPS
   docker ps | grep madabank
   curl http://localhost:8080/health
   ```

### Step 67: Jenkins Backup Script
Create backup script for Jenkins configuration.

`sudo nano /opt/bankingapp/scripts/backup-jenkins.sh`

```bash
#!/bin/bash
set -e

BACKUP_DIR="/opt/bankingapp/backups/jenkins"
JENKINS_HOME="/var/lib/jenkins"
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="jenkins-backup-${DATE}.tar.gz"

echo "Backing up Jenkins configuration..."

mkdir -p ${BACKUP_DIR}

# Backup important Jenkins directories
tar -czf ${BACKUP_DIR}/${BACKUP_FILE} \
    ${JENKINS_HOME}/config.xml \
    ${JENKINS_HOME}/credentials.xml \
    ${JENKINS_HOME}/jobs \
    ${JENKINS_HOME}/users \
    ${JENKINS_HOME}/secrets \
    ${JENKINS_HOME}/plugins \
    2>/dev/null || true

# Keep only last 7 backups
ls -t ${BACKUP_DIR}/jenkins-backup-*.tar.gz | tail -n +8 | xargs -r rm

echo "✅ Backup completed: ${BACKUP_FILE}"
ls -lh ${BACKUP_DIR}/${BACKUP_FILE}
```

```bash
sudo chmod +x /opt/bankingapp/scripts/backup-jenkins.sh
```

**Add to cron (weekly backup):**
```bash
crontab -e
# Add:
0 3 * * 0 /opt/bankingapp/scripts/backup-jenkins.sh >> /opt/bankingapp/logs/jenkins-backup.log 2>&1
```

### Jenkins CI/CD Workflow Diagram
```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Developer     │────▶│   GitHub Repo   │────▶│ GitHub Webhook  │
│   Push Code     │     │   (main branch) │     │  (POST event)   │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                         │
                                                         ▼
                                                ┌─────────────────┐
                                                │ Jenkins on VPS  │
                                                │ (Receives hook) │
                                                └────────┬────────┘
                                                         │
                        ┌────────────────────────────────┘
                        ▼
              ┌─────────────────┐
              │ Pipeline Stages │
              ├─────────────────┤
              │ 1. Checkout     │
              │ 2. Lint/Test    │
              │ 3. Build Binary │
              │ 4. Build Image  │
              │ 5. Deploy       │
              │ 6. Cleanup      │
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  ✅ Live at     │
              │  api.madabank   │
              └─────────────────┘
```

## ✅ Jenkins CI/CD Checkpoint
After completing Phase 10, you should have:
- ✅ Jenkins installed and running on VPS
- ✅ Jenkins secured behind Nginx with SSL
- ✅ GitHub webhook configured
- ✅ Jenkins pipeline job created
- ✅ Jenkinsfile in repository
- ✅ Automated build and deploy on push to `main`
- ✅ Jenkins backup configured

