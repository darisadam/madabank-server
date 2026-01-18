terraform {
  backend "s3" {
    bucket         = "madabank-terraform-state-dev"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "madabank-terraform-locks"
  }
}

module "madabank" {
  source = "../../"

  aws_region  = "us-east-1"
  environment = "prod"
  owner       = "darisadam.dev@gmail.com"

  # Networking
  vpc_cidr             = "10.2.0.0/16"
  public_subnet_cidrs  = ["10.2.1.0/24", "10.2.2.0/24", "10.2.3.0/24"]
  private_subnet_cidrs = ["10.2.11.0/24", "10.2.12.0/24", "10.2.13.0/24"]

  # Database - cost-optimized for learning/dev-prod
  db_instance_class       = "db.t3.micro"
  db_allocated_storage    = 20
  backup_retention_period = 7
  db_multi_az            = false # Single AZ to save cost

  # Redis - cost-optimized
  redis_node_type      = "cache.t3.micro"
  redis_num_cache_nodes = 1

  # ECS - cost-optimized
  container_cpu    = 256
  container_memory = 512
  desired_count    = 1
  min_capacity     = 1
  max_capacity     = 3

  # Container image
  container_image = "ghcr.io/darisadam/madabank-server:v1.0.0"

  # SSL Certificate
  certificate_arn = var.certificate_arn

  # Monitoring
  alert_email = "darisadam.dev@gmail.com"
}

output "alb_url" {
  value = module.madabank.alb_url
}