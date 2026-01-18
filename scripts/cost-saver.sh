#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status

ACTION=$1
ENV=$2

if [ -z "$ACTION" ] || [ -z "$ENV" ]; then
    echo "Usage: ./cost-saver.sh [start|stop] [dev|staging|prod]"
    exit 1
fi

# Production Safety Check
if [ "$ENV" == "prod" ] || [ "$ENV" == "production" ]; then
    if [ "$ACTION" == "stop" ]; then
        echo "⚠️  WARNING: You are about to STOP the PRODUCTION environment!"
        read -p "Are you sure? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Operation cancelled."
            exit 1
        fi
    fi
fi

CLUSTER="madabank-$ENV"
SERVICE="madabank-$ENV"

echo "🔍 Finding Resources for $ENV..."
# Fetch RDS Instance ID dynamically
RDS_ID=$(aws rds describe-db-instances --query "DBInstances[?contains(DBInstanceIdentifier, 'madabank-$ENV')].DBInstanceIdentifier" --output text)

if [ -z "$RDS_ID" ]; then
    echo "❌ Could not find RDS instance for environment: $ENV"
    exit 1
fi

if [ "$ACTION" == "stop" ]; then
    echo "🛑 Stopping Environment: $ENV"
    
    # 1. Scale Down ECS (Graceful Shutdown)
    echo "📉 Scaling ECS Service to 0..."
    aws ecs update-service --cluster $CLUSTER --service $SERVICE --desired-count 0 > /dev/null
    
    echo "⏳ Waiting for ECS tasks to drain (Graceful Shutdown)..."
    aws ecs wait services-stable --cluster $CLUSTER --services $SERVICE
    echo "✅ ECS tasks drained."

    # 2. Stop RDS (Only after ECS is down)
    echo "🛑 Stopping RDS Instance ($RDS_ID)..."
    aws rds stop-db-instance --db-instance-identifier $RDS_ID > /dev/null
    
    echo "✅ Environment paused safely. Savings active."

elif [ "$ACTION" == "start" ]; then
    echo "▶️ Starting Environment: $ENV"
    
    # 1. Start RDS
    echo "⚡ Starting RDS Instance ($RDS_ID)..."
    aws rds start-db-instance --db-instance-identifier $RDS_ID > /dev/null
    
    echo "⏳ Waiting for DB to be available..."
    aws rds wait db-instance-available --db-instance-identifier $RDS_ID
    echo "✅ Database is up."
    
    # 2. Start ECS
    echo "🚀 Scaling ECS Service to 1..."
    aws ecs update-service --cluster $CLUSTER --service $SERVICE --desired-count 1 > /dev/null
    
    echo "⏳ Waiting for Service to stabilize..."
    aws ecs wait services-stable --cluster $CLUSTER --services $SERVICE
    echo "✅ Environment started."
fi
