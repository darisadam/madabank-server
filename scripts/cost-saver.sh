#!/bin/bash

ACTION=$1
ENV=$2

if [ -z "$ACTION" ] || [ -z "$ENV" ]; then
    echo "Usage: ./cost-saver.sh [start|stop] [dev|staging|prod]"
    exit 1
fi

CLUSTER="madabank-$ENV"
SERVICE="madabank-$ENV"
# RDS ID might differ, fetching it for better reliability if strict naming wasn't used or TF generated suffix
# But we used 'madabank-prod' in our TF update, wait... TF random suffix?
# In dev/main.tf: module "rds" -> identifier = "madabank-dev-db" usually?
# Let's check TF. The module likely adds a suffix.
# Assuming standard naming for now: "madabank-$ENV-db" or similar.
# I'll use a dynamic lookup.

RDS_ID=$(aws rds describe-db-instances --query "DBInstances[?contains(DBInstanceIdentifier, 'madabank-$ENV')].DBInstanceIdentifier" --output text)

if [ "$ACTION" == "stop" ]; then
    echo "🛑 Stopping Environment: $ENV"
    
    echo "Scaling ECS Service to 0..."
    aws ecs update-service --cluster $CLUSTER --service $SERVICE --desired-count 0 > /dev/null
    
    echo "Stopping RDS Instance ($RDS_ID)..."
    aws rds stop-db-instance --db-instance-identifier $RDS_ID > /dev/null
    
    echo "✅ Environment paused. You are now saving money."

elif [ "$ACTION" == "start" ]; then
    echo "▶️ Starting Environment: $ENV"
    
    echo "Starting RDS Instance ($RDS_ID)..."
    aws rds start-db-instance --db-instance-identifier $RDS_ID > /dev/null
    
    echo "Waiting for DB to be available..."
    aws rds wait db-instance-available --db-instance-identifier $RDS_ID
    
    echo "Scaling ECS Service to 1..."
    aws ecs update-service --cluster $CLUSTER --service $SERVICE --desired-count 1 > /dev/null
    
    echo "✅ Environment started."
fi
