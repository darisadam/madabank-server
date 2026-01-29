# 🚀 Deployment Guide

This guide details the **Hybrid Deployment Strategy** for MadaBank, combining the cost-effectiveness of a Private VPS for Production with the scalability of AWS for Development/Staging.

## 🌍 Environments

| Environment | URL | Branch | Infrastructure | Orchestrator |
|-------------|-----|--------|----------------|--------------|
| **Development** | `api-dev.madabank.art` | `develop` | AWS (T3.micro) | ECS Fargate |
| **Staging** | `api-staging.madabank.art` | `staging` | AWS (T3.micro) | ECS Fargate |
| **Production** | `api.madabank.art` | `main` | Private VPS (Ubuntu 24) | Docker Compose + Jenkins |

---

## 🤖 CI/CD Strategy

We use a "GitOps-Hybrid" approach:

### 1. GitHub Actions (CI & Dev/Staging CD)
GitHub Actions handles the Continuous Integration (CI) for all branches and the Continuous Deployment (CD) for AWS environments.

*   **CI Checks**: runs `lint`, `test`, `build` on every Push/PR.
*   **Dev/Staging Deploy**: Pushes Docker image to GHCR -> Updates AWS ECS Service via Terraform/AWS CLI.

### 2. Jenkins (Production CD)
Jenkins runs on the private VPS and manages the Production deployment to ensure strict control and security within the private network.

*   **Trigger**: Push to `main`.
*   **Pipeline**:
    1.  Test & Build Docker Image.
    2.  Push to GHCR (`:latest`).
    3.  **Deployment**: Pulls the new image and re-ups the Docker Compose service.
    4.  **Release**: Tags the commit on GitHub.

👉 **[See Jenkins Setup Guide](JENKINS_SETUP.md)**

---

## 🛠️ Infrastructure Provisioning

### 1. Private VPS (Production)
We use Ansible or Shell Scripts to provision the bare metal VPS.

*   **Setup Script**: `scripts/vps/setup_vps.sh`
*   **Ansible Playbook**: `ansible/playbook.yml`

👉 **[See VPS Setup Guide](VPS_SETUP.md)**

### 2. AWS (Dev/Staging)
We use Terraform to manage the AWS infrastructure.

**Prerequisites:** Ubuntu/Mac with Terraform v1.5+ and AWS CLI.

```bash
# Apply Development Infrastructure
cd terraform/environments/dev
terraform init
terraform apply
```

---

## 🕵️ Troubleshooting

### Jenkins Deployment Failed
*   **Logs**: Check Jenkins Dashboard at `jenkins.madabank.art`.
*   **Docker**: SSH into VPS and check `docker logs madabank-api-prod`.

### AWS ECS Context Deadline Exceeded
*   **Cause**: Container cannot reach Internet/SecretsManager.
*   **Fix**: Check NAT Gateway status or Security Group Egress rules.
