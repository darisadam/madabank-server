#!/bin/bash

# Script to deploy infrastructure to specific environment
# Usage: ./scripts/deploy.sh <environment> [plan|apply|destroy]

set -e

ENVIRONMENT=$1
ACTION=${2:-plan}

if [ -z "$ENVIRONMENT" ]; then
  echo "Usage: $0 <environment> [plan|apply|destroy]"
  echo "Example: $0 dev plan"
  exit 1
fi

TERRAFORM_DIR="terraform/environments/$ENVIRONMENT"

if [ ! -d "$TERRAFORM_DIR" ]; then
  echo "Error: Environment directory not found: $TERRAFORM_DIR"
  exit 1
fi

echo "========================================="
echo "Terraform $ACTION for $ENVIRONMENT"
echo "========================================="
echo ""

cd $TERRAFORM_DIR

# Initialize Terraform
echo "Initializing Terraform..."
terraform init -upgrade

# Validate configuration
echo "Validating configuration..."
terraform validate

# Format check
echo "Checking formatting..."
terraform fmt -check -recursive || {
  echo "Warning: Code is not formatted. Run 'terraform fmt -recursive'"
}

case $ACTION in
  plan)
    echo "Creating execution plan..."
    terraform plan -out=tfplan
    echo ""
    echo "✅ Plan created successfully!"
    echo "To apply: ./scripts/deploy.sh $ENVIRONMENT apply"
    ;;
    
  apply)
    if [ ! -f "tfplan" ]; then
      echo "Error: No plan file found. Run plan first."
      exit 1
    fi
    
    echo "Applying changes..."
    terraform apply tfplan
    rm -f tfplan
    
    echo ""
    echo "✅ Infrastructure deployed successfully!"
    terraform output
    ;;
    
  destroy)
    echo "⚠️  WARNING: This will destroy all infrastructure in $ENVIRONMENT!"
    echo "Type 'yes' to continue:"
    read -r confirmation
    
    if [ "$confirmation" != "yes" ]; then
      echo "Aborted."
      exit 0
    fi
    
    terraform destroy
    echo "✅ Infrastructure destroyed."
    ;;
    
  *)
    echo "Invalid action: $ACTION"
    echo "Valid actions: plan, apply, destroy"
    exit 1
    ;;
esac