# 💰 AWS Cost Management Guide

To prevent your AWS credits from draining, you can "pause" your environment when not in use. You **DO NOT** need to destroy everything (which deletes data).

## Option 1: The "Pause" Button (Recommended)

This stops the compute resources (ECS & RDS) so you stop paying hourly rates, but keeps your data (Storage/EBS) and networking (ALB/VPC).

### 🛑 STOP Everything
Run this command (or script provided below):
```bash
# 1. Scale ECS Service to 0 (Stops containers)
aws ecs update-service --cluster madabank-prod --service madabank-prod --desired-count 0

# 2. Stop RDS Database (Stops compute)
aws rds stop-db-instance --db-instance-identifier madabank-prod
```
*Note: RDS will automatically restart after 7 days (AWS rule).*

### ▶️ START Everything
```bash
# 1. Start RDS Database
aws rds start-db-instance --db-instance-identifier madabank-prod

# 2. Wait a few minutes, then Scale ECS Service to 1 (or 3)
aws ecs update-service --cluster madabank-prod --service madabank-prod --desired-count 1
```

## Option 2: The "Nuke" Button (Terraform Destroy)
**⚠️ WARNING: DELETES ALL DATA IN DATABASE**

Only do this if you are done with the project or have backups.
```bash
cd terraform/environments/prod
terraform destroy
```

## Option 3: Use the Helper Script
I have created a script `scripts/cost-saver.sh` for you.

```bash
# Stop everything
./scripts/cost-saver.sh stop prod

# Start everything
./scripts/cost-saver.sh start prod
```
