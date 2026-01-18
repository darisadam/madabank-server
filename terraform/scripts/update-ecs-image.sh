#!/bin/bash

# Script to update ECS service with new container image
# Usage: ./scripts/update-ecs-image.sh <environment> <image-tag>

set -e

ENVIRONMENT=$1
IMAGE_TAG=$2

if [ -z "$ENVIRONMENT" ] || [ -z "$IMAGE_TAG" ]; then
  echo "Usage: $0 <environment> <image-tag>"
  echo "Example: $0 staging v1.2.3"
  exit 1
fi

CLUSTER_NAME="madabank-${ENVIRONMENT}"
SERVICE_NAME="madabank-${ENVIRONMENT}"
IMAGE="ghcr.io/darisadam/madabank-server:${IMAGE_TAG}"

echo "========================================="
echo "Updating ECS Service"
echo "Environment: $ENVIRONMENT"
echo "Image: $IMAGE"
echo "========================================="
echo ""

# Get current task definition
echo "Fetching current task definition..."
TASK_DEFINITION=$(aws ecs describe-services \
  --cluster $CLUSTER_NAME \
  --services $SERVICE_NAME \
  --query 'services[0].taskDefinition' \
  --output text)

echo "Current task definition: $TASK_DEFINITION"

# Get task definition JSON
TASK_DEF_JSON=$(aws ecs describe-task-definition \
  --task-definition $TASK_DEFINITION \
  --query 'taskDefinition')

# Update image in task definition
NEW_TASK_DEF=$(echo $TASK_DEF_JSON | jq --arg IMAGE "$IMAGE" \
  '.containerDefinitions[0].image = $IMAGE | 
   del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy)')

# Register new task definition
echo "Registering new task definition..."
NEW_TASK_DEF_ARN=$(echo $NEW_TASK_DEF | \
  aws ecs register-task-definition \
  --cli-input-json file:///dev/stdin \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)

echo "New task definition: $NEW_TASK_DEF_ARN"

# Update service
echo "Updating ECS service..."
aws ecs update-service \
  --cluster $CLUSTER_NAME \
  --service $SERVICE_NAME \
  --task-definition $NEW_TASK_DEF_ARN \
  --force-new-deployment

echo ""
echo "✅ Deployment initiated!"
echo "Monitor deployment: aws ecs describe-services --cluster $CLUSTER_NAME --services $SERVICE_NAME"