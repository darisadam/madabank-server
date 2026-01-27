# GitHub Repository Configuration Guide

To enable the new CI/CD flow, you need to configure the following in your GitHub Repository Settings.

## 1. Secrets (Settings -> Secrets and variables -> Actions)
You mentioned you deleted everything. Please add these back:

| Secret Name | Value Description | Used By |
| :--- | :--- | :--- |
| `GHCR_USERNAME` | Your GitHub Username (`darisadam`) | Jenkins / GitHub Actions |
| `GHCR_PAT` | A Personal Access Token with `write:packages` scope | Jenkins / GitHub Actions |
| `VPS_HOST` | The IP Address of your VPS | GitHub Actions (if using SSH) |
| `VPS_USER` | The username on VPS (e.g. `root` or `madabank`) | GitHub Actions (if using SSH) |
| `VPS_SSH_KEY` | The Private SSH Key (contents of `.pem` or `id_rsa`) | GitHub Actions (if using SSH) |

## 2. Environments (Settings -> Environments)
Create these environments to add protection rules (optional but good practice):
1.  **development**
2.  **staging**
3.  **production**

## 3. Jenkins Credentials (In Jenkins UI)
Since Jenkins is doing the heavy lifting, you need to add these **inside Jenkins**:
1.  **Credentials ID:** `github-registry-credentials`
    *   **Type:** Username with Password
    *   **Username:** `darisadam`
    *   **Password:** Your GitHub PAT
2.  **Credentials ID:** `github-git-creds`
    *   **Type:** Username with Password
    *   **Username:** `darisadam`
    *   **Password:** Your GitHub PAT (needs `repo` scope to push tags)
3.  **Credentials ID:** `madabank-env-prod` (Secret File)
    *   Upload your production `.env` file.

## 4. Webhook Setup
To trigger Jenkins on Push/PR:
1.  Go to GitHub Repo -> Settings -> Webhooks.
2.  Add Webhook: `http://<YOUR_VPS_IP>/github-webhook/`
3.  Content type: `application/json`
4.  Select "Just the push event" (or "Let me select individual events" -> Pushes & Pull Requests).
