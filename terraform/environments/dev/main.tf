terraform {
  backend "s3" {
    bucket         = "madabank-terraform-state-dev"
    key            = "dev/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "madabank-terraform-locks"
  }
}

provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = {
      Project     = "MadaBank"
      Environment = "dev"
      ManagedBy   = "Terraform"
      Owner       = "darisadam.dev@gmail.com"
    }
  }
}

variable "container_image" {
  description = "Docker image tag to deploy"
  type        = string
}

variable "docker_username" {
  description = "CI/CD Docker username"
  type        = string
  sensitive   = true
}

variable "docker_password" {
  description = "CI/CD Docker password"
  type        = string
  sensitive   = true
}

module "madabank" {
  source = "../../modules/application-stack"

  aws_region  = "us-east-1"
  environment = "dev"
  owner       = "darisadam.dev@gmail.com"
  
  # Networking
  vpc_cidr             = "10.0.0.0/16"
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24"]
  single_nat_gateway   = true

  # Database - small instance for dev
  db_instance_class       = "db.t3.micro"
  db_allocated_storage    = 20
  backup_retention_period = 1
  db_multi_az             = false
  db_name                 = "madabank"
  db_username             = "madabankadmin"

  # Redis - small instance for dev
  redis_node_type       = "cache.t3.micro"
  redis_num_cache_nodes = 1

  # ECS - minimal resources for dev
  container_cpu    = 256
  container_memory = 512
  desired_count    = 1
  min_capacity     = 1
  max_capacity     = 2

  # Monitoring
  alert_email = "darisadam.dev@gmail.com"

  # Container Image (Passed from CI/CD)
  container_image = var.container_image

  # Docker Credentials
  docker_username = var.docker_username
  docker_password = var.docker_password

  # ALB (Placeholder ARN for dev)
  certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/placeholder"
}

output "alb_url" {
  value = module.madabank.alb_url
}

output "alb_dns_name" {
  value = module.madabank.alb_dns_name
}