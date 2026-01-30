# 3. Monitoring Setup Guide

**Goal:** Setup observability using Prometheus (Metrics), Grafana (Dashboard), and Loki (Logs).

## Stack Overview
- **Prometheus**: Collects metrics from API (`/metrics`), Node Exporter (System), and Postgres Exporter (DB).
- **Grafana**: Visualizes metrics and logs.
- **Loki**: Aggregates logs from all containers.
- **Promtail**: Ships logs from Docker to Loki.

## Configuration

### 1. Prometheus
Config file: `monitoring/prometheus/prometheus.yml`
Ensure targets are correctly defined:
```yaml
scrape_configs:
  - job_name: 'api'
    static_configs:
      - targets: ['madabank-api:8080']
```

### 2. Grafana
- **URL**: `https://monitoring.madabank.art`
- **Default Login**: `admin` / `admin` (Change immediately!)
- **Data Sources**:
    - Prometheus: `http://prometheus:9090`
    - Loki: `http://loki:3100`

### 3. Deployment
The monitoring stack is part of the main `docker-compose.yml`.
To deploy/update only monitoring:
```bash
docker compose up -d prometheus grafana loki promtail node-exporter
```
