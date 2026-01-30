# 2. CI/CD Setup Guide

**Goal:** Configure Continuous Integration (GitHub Actions) and Continuous Deployment (Jenkins).

## Part A: GitHub Actions (CI)
GitHub Actions handles **Linting**, **Testing**, and **Security Scanning**.

1.  **Workflow File**: `.github/workflows/ci.yml` must exist in your repository.
2.  **Triggers**: Pushes to `main`, `develop`, `staging` and Pull Requests.

No VPS configuration is required for GitHub Actions unless you are using self-hosted runners (default uses `ubuntu-latest`).

---

## Part B: Jenkins (CD)
Jenkins runs on the VPS and handles **Building Docker Images** and **Deployment**.

### 1. Installation
Jenkins runs in a Docker container (see `docker-compose.yml`):
- **Port**: `8080` (Internal), `8081` (Mapped to host localhost)
- **Volume**: `jenkins-data`

### 2. Access Jenkins
Tunnel port 8081 to your local machine:
```bash
ssh -L 8081:localhost:8081 admin@your-vps-ip
```
Open `http://localhost:8081` in your browser.

### 3. Configuration
1.  **Install Plugins**:
    - `Docker`, `Docker Pipeline`, `Pipeline`, `Git`.
2.  **Add Credentials**:
    - **ID**: `github-git-creds`
    - **Username**: Your GitHub Username
    - **Password**: GitHub PAT (scopes: `repo`, `read:packages`).

### 4. Create Pipeline
- **Type**: Multibranch Pipeline
- **Source**: GitHub
- **Repo**: `https://github.com/darisadam/madabank-server.git`
- **Script Path**: `Jenkinsfile`
