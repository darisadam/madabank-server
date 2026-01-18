# 🚀 Deployment Guide

This guide details the deployment strategy for MadaBank, covering both Automated CI/CD pipelines and Manual Infrastructure provisioning.

## 🌍 Environments

| Environment | URL | Branch | Infrastructure |
|-------------|-----|--------|----------------|
| **Development** | `api-dev.madabank.art` | `develop` | T3.micro / Single AZ |
| **Staging** | `api-staging.madabank.art` | `staging` | T3.micro / Single AZ |
| **Production** | `api.madabank.art` | `main` | Production Grade / Multi-AZ Capable |

---

## 🤖 CI/CD Pipelines (GitHub Actions)

We use a "GitOps-lite" approach where pushing to specific branches triggers deployments.

### 1. Development Deployment
*   **Trigger**: Push to `develop`.
*   **Action**: Builds Docker image, pushes to GHCR, updates ECS Service `madabank-dev`.

### 2. Staging Deployment
*   **Trigger**: Push to `staging`.
*   **Action**: Builds Docker image, pushes to GHCR, updates ECS Service `madabank-staging`.

### 3. Production Deployment
*   **Trigger**: Push to `main`.
*   **Action**: Builds Docker image, pushes to GHCR with `latest` tag, updates ECS Service `madabank-prod`.
*   **Strategy**: Uses Rolling Update (Min 100%, Max 200%) for zero-downtime.

---

## 🏗️ Infrastructure Provisioning (Terraform)

If building the infrastructure from scratch (or disaster recovery), follow these steps.

### Prerequisites
*   Terraform v1.5+
*   AWS CLI configured with Admin credentials
*   S3 Bucket for Terraform State (Created manually or via bootstrap script)

### Step 1: Initialize & Apply Development
```bash
cd terraform/environments/dev
terraform init
terraform apply -var="docker_password=YOUR_GH_TOKEN"
```

### Step 2: Initialize & Apply Staging
```bash
cd terraform/environments/staging
terraform init
terraform apply -var="docker_password=YOUR_GH_TOKEN"
```

### Step 3: Initialize & Apply Production
```bash
cd terraform/environments/prod
terraform init
terraform apply -var="docker_password=YOUR_GH_TOKEN"
```

---

## 🕵️ Troubleshooting

### "Context Deadline Exceeded" (ECS)
*   **Cause**: Container cannot reach Internet/SecretsManager.
*   **Fix**: Check NAT Gateway status or Security Group Egress rules. Ensure `single_nat_gateway` is configured correctly for cost savings.

### "AddressLimitExceeded" (EIP)
*   **Cause**: You hit the limit of 5 Elastic IPs.
*   **Fix**: Ensure `dev` and `staging` use `single_nat_gateway = true` to save IPs. Use the `scripts/manage-*.sh` tools to verify environment state.
