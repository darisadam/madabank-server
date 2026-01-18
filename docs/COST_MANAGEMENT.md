# 💰 AWS Cost Management Guide

To prevent your AWS credits from draining, you can "pause" your environment when not in use. You **DO NOT** need to destroy everything (which deletes data).

## 🚀 Quick Management Scripts (Recommended)

We have provided wrapper scripts for each environment to make this easy. These scripts perform a **graceful shutdown** (draining ECS connections first) before stopping the database.

### 🏠 Development
```bash
# Stop Dev
./scripts/manage-dev.sh stop

# Start Dev
./scripts/manage-dev.sh start
```

### 🧪 Staging
```bash
# Stop Staging
./scripts/manage-staging.sh stop

# Start Staging
./scripts/manage-staging.sh start
```

### 🏭 Production
```bash
# Stop Production (Requires Confirmation)
./scripts/manage-prod.sh stop

# Start Production
./scripts/manage-prod.sh start
```

### 🌍 All Environments (Global)
Manage everything at once (useful for "end of day" shutdown).

```bash
# Stop ALL environments (Dev, Staging, Prod)
./scripts/manage-all.sh stop

# Start ALL environments
./scripts/manage-all.sh start
```

---

## ⚙️ How it Works (Under the Hood)

These scripts call `scripts/cost-saver.sh`, which performs the following steps:

1.  **Identify Resources**: Finds the RDS Instance and ECS Services for the target environment.
2.  **Stop Sequence**:
    *   Scales ECS Services to `0` desired tasks.
    *   **Waits** for services to stabilize (draining connections).
    *   Stops the RDS Database Instance.
3.  **Start Sequence**:
    *   Starts the RDS Database Instance.
    *   **Waits** for the DB to be available.
    *   Scales ECS Services back to `1` (or original count).
    *   Waits for services to stabilize.

## 🗑️ Option 2: The "Nuke" Button (Terraform Destroy)

**⚠️ WARNING: DELETES ALL DATA IN DATABASE PERMANENTLY**

Only do this if you are done with the project or want to rebuild from scratch.

```bash
cd terraform/environments/prod
terraform destroy
```
