# MadaBank Terraform Infrastructure

This directory contains Infrastructure as Code (IaC) for deploying MadaBank to AWS using Terraform.

## Architecture Overview
```
┌─────────────────────────────────────────────────────────────┐
│                         Internet                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
                 ┌───────────────┐
                 │ Application   │
                 │ Load Balancer │
                 │ (Public)      │
                 └───────┬───────┘
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
    ┌────────┐     ┌────────┐     ┌────────┐
    │ ECS    │     │ ECS    │     │ ECS    │
    │ Task 1 │     │ Task 2 │     │ Task 3 │
    └────┬───┘     └────┬───┘     └────┬───┘
         │              │              │
         └──────────────┼──────────────┘
                        │
         ┌──────────────┼──────────────┐
         ▼              ▼              ▼
    ┌─────────┐   ┌─────────┐   ┌─────────┐
    │   RDS   │   │  Redis  │   │ Secrets │
    │ (Multi  │   │(ElastiC)│   │ Manager │
    │   AZ)   │   │         │   │         │
    └─────────┘   └─────────┘   └─────────┘
```

## Directory Structure
```
terraform/
├── main.tf                 # Root module
├── variables.tf            # Root variables
├── outputs.tf              # Root outputs
├── modules/                # Reusable modules
│   ├── networking/         # VPC, subnets, NAT
│   ├── security/           # Security groups
│   ├── iam/                # IAM roles and policies
│   ├── rds/                # PostgreSQL database
│   ├── elasticache/        # Redis cluster
│   ├── alb/                # Application Load Balancer
│   ├── ecs/                # ECS cluster and service
│   └── monitoring/         # CloudWatch alarms
├── environments/           # Environment configurations
│   ├── dev/
│   ├── staging/
│   └── prod/
└── scripts/                # Helper scripts
    ├── init-backend.sh
    ├── deploy.sh
    └── update-ecs-image.sh
```

## Prerequisites

1. **AWS CLI** configured with credentials
2. **Terraform** >= 1.0 installed
3. **jq** for JSON processing (for update scripts)
```bash
# Install Terraform
brew install terraform  # macOS
# or download from https://www.terraform.io/downloads

# Configure AWS CLI
aws configure
```

## Quick Start

### 1. Initialize Backend

Create S3 bucket and DynamoDB table for Terraform state:
```bash
cd terraform
./scripts/init-backend.sh dev
```

### 2. Deploy Infrastructure
```bash
# Plan deployment
./scripts/deploy.sh dev plan

# Apply changes
./scripts/deploy.sh dev apply

# Get outputs
cd environments/dev
terraform output
```

### 3. Access Your Application

After deployment, get the ALB DNS name:
```bash
terraform output alb_url
```

Visit: `http://<alb-dns-name>`

## Cost Estimates

### Development Environment (~$30/month)
- RDS (db.t3.micro): ~$15/month
- ElastiCache (cache.t3.micro): ~$12/month
- ECS Fargate (1 task): ~$8/month
- ALB: ~$16/month
- NAT Gateway: ~$32/month
- **Total: ~$83/month**

### Production Environment (~$200/month)
- RDS (db.t3.medium, Multi-AZ): ~$150/month
- ElastiCache (cache.t3.medium, 2 nodes): ~$100/month
- ECS Fargate (3 tasks): ~$75/month
- ALB: ~$16/month
- NAT Gateway (2): ~$64/month
- **Total: ~$405/month**

## Environments

### Dev
- Minimal resources
- No Multi-AZ
- 1-day backups
- 1 ECS task

### Staging
- Medium resources
- No Multi-AZ
- 7-day backups
- 2 ECS tasks

### Production
- Production-grade resources
- Multi-AZ enabled
- 30-day backups
- 3+ ECS tasks (autoscaling)

## Deployment

### Deploy New Application Version
```bash
# After building and pushing Docker image
./scripts/update-ecs-image.sh staging v1.2.3
```

### Update Infrastructure
```bash
# Make changes to Terraform files
cd environments/staging

# Plan and review
terraform plan

# Apply changes
terraform apply
```

### Rollback
```bash
# Revert to previous task definition
aws ecs update-service \
  --cluster madabank-staging \
  --service madabank-staging \
  --task-definition madabank-staging:<previous-revision>
```

## Monitoring

### CloudWatch Dashboard
After deployment, access your CloudWatch dashboard:
```bash
terraform output dashboard_url
```

### Logs
View ECS logs:
```bash
aws logs tail /ecs/madabank-dev --follow
```

### Metrics
Key metrics to monitor:
- ECS CPU/Memory utilization
- ALB target response time
- RDS connections
- Redis memory usage

## Troubleshooting

### ECS Tasks Not Starting
```bash
# Check service events
aws ecs describe-services \
  --cluster madabank-dev \
  --services madabank-dev \
  --query 'services[0].events'

# Check task logs
aws logs tail /ecs/madabank-dev --follow
```

### Database Connection Issues
```bash
# Verify security groups
aws ec2 describe-security-groups \
  --group-ids <ecs-sg-id>

# Test from ECS task
aws ecs execute-command \
  --cluster madabank-dev \
  --task <task-id> \
  --container madabank-api \
  --command "/bin/sh" \
  --interactive
```

### High Costs

1. **NAT Gateway**: Most expensive component
   - Consider using VPC endpoints for S3/DynamoDB
   - Use single NAT in non-prod environments

2. **RDS**: Right-size your instance
   - Start with t3.micro in dev
   - Monitor actual usage and scale accordingly

3. **ECS**: Optimize task resources
   - Reduce CPU/memory if underutilized
   - Use Fargate Spot for non-critical workloads

## Security

### Secrets Management
All secrets stored in AWS Secrets Manager:
- Database password
- JWT secret
- Encryption keys

### Network Security
- Private subnets for application and database
- Security groups with least privilege
- VPC Flow Logs enabled

### Encryption
- RDS encryption at rest
- ElastiCache encryption at rest
- S3 encryption for logs

## Cleanup

**⚠️ WARNING: This will delete all resources and data!**
```bash
./scripts/deploy.sh dev destroy
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/darisadam/madabank-server/issues
- Email: darisadam.dev@gmail.com