# 4. Application Deployment Guide

**Goal:** Configure and deploy the core application, database, and reverse proxy.

## Phase 1: Environment Setup

### 1. Generate Secrets
Run this on your local machine and **save the output securely**.
```bash
# Database & Redis Passwords
openssl rand -base64 32

# JWT Secrets (64 chars)
openssl rand -base64 64

# Encryption Key (32 chars)
openssl rand -base64 32
```

### 2. Configure .env
On VPS: `nano /opt/madabankapp/.env`
> See `.env.example` in the repo for the template.

**Critical Variables:**
- `DB_PASSWORD`: Use the generated DB password.
- `ENCRYPTION_KEY`: Must be exactly 32 raw characters (or 44 chars if Base64).
- `JWT_SECRET`: Must be long and secure.

## Phase 2: Nginx Reverse Proxy

### 1. Configuration
File: `docker/nginx/conf.d/api.conf`
Ensure the upstream points to the correct port:
```nginx
location / {
    proxy_pass http://madabank-api:3000;
}
```

### 2. SSL Certificates
We use Certbot for SSL.
```bash
# Initial Run (Standalone mode)
docker run -it --rm --name certbot \
    -v "/opt/madabankapp/certbot/conf:/etc/letsencrypt" \
    -v "/opt/madabankapp/certbot/www:/var/www/certbot" \
    certbot/certbot certonly --standalone -d api.madabank.art
```

## Phase 3: Start Services

```bash
cd /opt/madabankapp/docker
docker compose up -d
```

### Verification
```bash
# Check status
docker compose ps

# Check API health
curl -k https://localhost/health
```
