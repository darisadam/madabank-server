# MadaBank API

[![CI Pipeline](https://github.com/darisadam/madabank-server/actions/workflows/ci.yml/badge.svg)](https://github.com/darisadam/madabank-server/actions/workflows/ci.yml)
[![CD Pipeline](https://github.com/darisadam/madabank-server/actions/workflows/cd.yml/badge.svg)](https://github.com/darisadam/madabank-server/actions/workflows/cd.yml)
[![codecov](https://codecov.io/gh/darisadam/madabank-server/branch/main/graph/badge.svg)](https://codecov.io/gh/darisadam/madabank-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/darisadam/madabank-server)](https://goreportcard.com/report/github.com/darisadam/madabank-server)

> Production-grade banking monolith demonstrating DevOps excellence

## 🎯 Project Goals

This project demonstrates enterprise-level backend and DevOps practices:
- ✅ ACID-compliant financial transactions
- ✅ Security-first architecture (encryption, JWT, secrets management)
- ✅ Full observability (metrics, logs, traces, alerts)
- ✅ Automated CI/CD with security scanning
- ✅ Cost-optimized AWS deployment
- ✅ Kubernetes-ready architecture
- ✅ ISO 27001 & CMMI compliance concepts

## 🚀 Quick Start
```bash
# Clone repository
git clone https://github.com/darisadam/madabank-server.git
cd madabank-server

# Run with Docker Compose
make docker-up

# Run tests
make test

# View coverage
make test-coverage
```

## 📊 CI/CD Pipeline

Our automated pipeline includes:
- **Linting & Code Quality**: golangci-lint, gofmt, go vet
- **Testing**: Unit tests with 70%+ coverage
- **Security Scanning**: Gosec, Trivy, Nancy
- **Docker Build**: Multi-stage optimized builds
- **Automated Deployment**: ECS Fargate (staging/production)

### Running CI Checks Locally
```bash
# Lint code
make lint

# Run all tests
make test

# Security scan
make security-scan

# Build Docker image
make docker-build
```

## 🛡️ Security

Security is a top priority:
- All passwords hashed with bcrypt
- JWT authentication with RS256
- Encryption at rest (AES-256-GCM)
- TLS/HTTPS enforced
- SQL injection protection
- Rate limiting
- Audit logging for all operations

See [SECURITY.md](docs/SECURITY.md) for details.

## 📚 Documentation

- [API Documentation](docs/API.md)
- [Architecture Overview](docs/ARCHITECTURE.md)
- [Security Model](docs/SECURITY.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Contributing Guidelines](CONTRIBUTING.md)

## 🛠️ Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.24+ | High performance, type safety |
| **Framework** | Gin | Fast HTTP routing |
| **Database** | PostgreSQL 16 | ACID compliance |
| **Cache** | Redis 7 | Session & rate limiting |
| **Container** | Docker | Portability |
| **Orchestration** | ECS Fargate | Cost-effective deployment |
| **IaC** | Terraform | Infrastructure automation |
| **CI/CD** | GitHub Actions | Automated pipelines |
| **Monitoring** | Prometheus + Grafana | Observability |
| **Security** | Gosec, Trivy | Vulnerability scanning |

## 🧪 Testing
```bash
# Unit tests
go test -v ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests
go test -v ./tests/integration/...

# Benchmark tests
go test -bench=. -benchmem ./...
```

## 🚀 Deployment

This project uses a **Tuple Deployment Strategy** with fully isolated environments:

| Environment | Branch | URL | Purpose |
|-------------|--------|-----|---------|
| **Development** | `develop` | `https://api-dev.madabank.art` | Frontend Integration Testing |
| **Staging** | `staging` | `https://api-staging.madabank.art` | QA & Acceptance Testing |
| **Production** | `main` | `https://api.madabank.art` | Live User Traffic |

### 💰 Cost Management (New)
To prevent AWS bill shock, we have implemented a "shutdown" mechanism for non-production hours.
See [Cost Management Guide](docs/COST_MANAGEMENT.md).

### Deployment Commands

#### Staging (Manual Trigger via Git)
```bash
git checkout staging
git merge develop
git push origin staging
```

#### Production (Automatic)
```bash
git checkout main
git merge staging
git push origin main
```

#### Production Release (Tagging)
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## 📈 Roadmap

- [x] User authentication & authorization (JWT + Refresh Tokens)
- [x] Account management
- [x] Transaction system with ACID compliance
- [x] CI/CD pipeline (GitHub Actions -> AWS ECS)
- [x] AWS Infrastructure (Terraform for Dev/Staging/Prod)
- [x] Rate limiting & DDoS protection
- [x] Maintenance Mode
- [x] Card management encryptions (AES-256 + RSA-2048)
- [x] Prometheus metrics & Grafana dashboards
- [x] iOS mobile app integration (API Ready)

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md).

## 👤 Author

**Daris Adam**
- GitHub: [@darisadam](https://github.com/darisadam)

---

**Status:** ✅ Production Ready