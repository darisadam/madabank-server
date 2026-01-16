#!/bin/bash

# Script to initialize Terraform S3 backend and DynamoDB table
# Usage: ./scripts/init-backend.sh <environment>

set -e

ENVIRONMENT=$1

if [ -z "$ENVIRONMENT" ]; then
  echo "Usage: $0 <environment>"
  echo "Example: $0 dev"
  exit 1
fi

AWS_REGION="us-east-1"
BUCKET_NAME="madabank-terraform-state-${ENVIRONMENT}"
DYNAMODB_TABLE="madabank-terraform-locks"

echo "========================================="
echo "Initializing Terraform Backend"
echo "Environment: $ENVIRONMENT"
echo "========================================="
echo ""

# Create S3 bucket for Terraform state
echo "Creating S3 bucket: $BUCKET_NAME"
aws s3api create-bucket \
  --bucket $BUCKET_NAME \
  --region $AWS_REGION || echo "Bucket already exists"

# Enable versioning
echo "Enabling versioning..."
aws s3api put-bucket-versioning \
  --bucket $BUCKET_NAME \
  --versioning-configuration Status=Enabled

# Enable encryption
echo "Enabling encryption..."
aws s3api put-bucket-encryption \
  --bucket $BUCKET_NAME \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Block public access
echo "Blocking public access..."
aws s3api put-public-access-block \
  --bucket $BUCKET_NAME \
  --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true

# Create DynamoDB table for state locking (only once)
echo ""
echo "Creating DynamoDB table: $DYNAMODB_TABLE"
aws dynamodb create-table \
  --table-name $DYNAMODB_TABLE \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --region $AWS_REGION || echo "Table already exists"

echo ""
echo "✅ Backend initialized successfully!"
echo ""
echo "Next steps:"
echo "1. cd terraform/environments/$ENVIRONMENT"
echo "2. terraform init"
echo "3. terraform plan"
echo "4. terraform apply"