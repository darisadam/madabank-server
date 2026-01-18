#!/bin/bash
set -e

echo "🚀 Starting Manual Deployment Fix..."

# 1. Check Docker
if ! docker info > /dev/null 2>&1; then
  echo "❌ Docker is not running or unresponsive."
  echo "👉 Please restart Docker Desktop and run this script again."
  exit 1
fi

# 2. Build Image (ARM64 for cost savings & match local arch)
echo "📦 Building Docker image (linux/arm64)..."
# Using -f docker/Dockerfile as discovered
docker build --platform linux/arm64 -t ghcr.io/darisadam/madabank-server:latest -f docker/Dockerfile .

# 3. Push Image
echo "Rx Pushing image to GHCR..."
docker push ghcr.io/darisadam/madabank-server:latest

# 4. Trigger ECS Deployment
echo "🔄 Updating ECS Service to pull new image..."
# Force new deployment to pick up the :latest tag we just pushed
aws ecs update-service --cluster madabank-dev --service madabank-dev --force-new-deployment > /dev/null

echo "✅ Deployment triggered!"
echo "The application should be running in a few minutes."
echo "You can check status with: aws ecs describe-services --cluster madabank-dev --services madabank-dev"
