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
