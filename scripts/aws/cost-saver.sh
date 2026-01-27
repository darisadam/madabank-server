#!/bin/bash
set -e

ACTION=$1
ENV=$2

if [ -z "$ACTION" ] || [ -z "$ENV" ]; then
    echo "Usage: ./cost-saver.sh [start|stop] [dev|staging|prod]"
    exit 1
fi

# Production Safety Check
if [[ "$ENV" == "prod" || "$ENV" == "production" ]]; then
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

echo "🔍 Finding Resources for $ENV (Cluster: $CLUSTER)..."

# 1. Fetch RDS Instance ID
RDS_ID=$(aws rds describe-db-instances --query "DBInstances[?contains(DBInstanceIdentifier, 'madabank-$ENV')].DBInstanceIdentifier" --output text)

if [ -z "$RDS_ID" ]; then
    echo "❌ Could not find RDS instance for environment: $ENV"
    exit 1
fi

# 2. Fetch All Services in Cluster
SERVICES=$(aws ecs list-services --cluster "$CLUSTER" --query "serviceArns[]" --output text)
if [ -z "$SERVICES" ]; then
    echo "⚠️  No services found in cluster $CLUSTER."
fi

if [ "$ACTION" == "stop" ]; then
    echo "🛑 Stopping Environment: $ENV"
    
    # 1. Scale Down ALL ECS Services
    if [ -n "$SERVICES" ]; then
        echo "📉 Scaling down services..."
        for SVC_ARN in $SERVICES; do
            SVC_NAME=${SVC_ARN##*/}
            echo "   - Stopping service: $SVC_NAME"
            aws ecs update-service --cluster "$CLUSTER" --service "$SVC_NAME" --desired-count 0 > /dev/null
        done
        
        echo "⏳ Waiting for ECS tasks to drain (Graceful Shutdown)..."
        # Wait for all services to stabilize
        # shellcheck disable=SC2086
        aws ecs wait services-stable --cluster "$CLUSTER" --services $SERVICES
        echo "✅ ECS tasks drained."
    fi

    # 2. Stop RDS (Only after ECS is down)
    echo "🛑 Stopping RDS Instance ($RDS_ID)..."
    aws rds stop-db-instance --db-instance-identifier "$RDS_ID" > /dev/null
    
    echo "✅ Environment paused safely. Savings active."

elif [ "$ACTION" == "start" ]; then
    echo "▶️ Starting Environment: $ENV"
    
    # 1. Start RDS first
    STATUS=$(aws rds describe-db-instances --db-instance-identifier "$RDS_ID" --query "DBInstances[0].DBInstanceStatus" --output text)
    if [ "$STATUS" == "stopped" ]; then
        echo "⚡ Starting RDS Instance ($RDS_ID)..."
        aws rds start-db-instance --db-instance-identifier "$RDS_ID" > /dev/null
        
        echo "⏳ Waiting for DB to be available..."
        aws rds wait db-instance-available --db-instance-identifier "$RDS_ID"
        echo "✅ Database is up."
    else
        echo "ℹ️  RDS Instance is already $STATUS."
    fi
    
    # 2. Start ECS Services
    if [ -n "$SERVICES" ]; then
        echo "🚀 Scaling up services..."
        for SVC_ARN in $SERVICES; do
            SVC_NAME=${SVC_ARN##*/}
            echo "   - Starting service: $SVC_NAME"
            aws ecs update-service --cluster "$CLUSTER" --service "$SVC_NAME" --desired-count 1 > /dev/null
        done
        
        echo "⏳ Waiting for Services to stabilize..."
        # shellcheck disable=SC2086
        aws ecs wait services-stable --cluster "$CLUSTER" --services $SERVICES
        echo "✅ Environment started."
    fi
fi
