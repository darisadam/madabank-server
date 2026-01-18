terraform {
  backend "s3" {
    bucket         = "madabank-terraform-state-dev"
    key            = "dev/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "madabank-terraform-locks"
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
  source = "../../"

  aws_region  = "us-east-1"
  environment = "dev"
  owner       = "darisadam.dev@gmail.com"

  # Networking
  vpc_cidr             = "10.0.0.0/16"
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24"]

  # Database - small instance for dev
  db_instance_class       = "db.t3.micro"
  db_allocated_storage    = 20
  backup_retention_period = 1
  db_multi_az             = false

  # Redis - small instance for dev
  redis_node_type       = "cache.t3.micro"
  redis_num_cache_nodes = 1

  # ECS - minimal resources for dev
  container_cpu    = 256
  container_memory = 512
  desired_count    = 1

  # Monitoring
  alert_email = "darisadam.dev@gmail.com"

  # Container Image (Passed from CI/CD)
  container_image = var.container_image

  # Docker Credentials
  docker_username = var.docker_username
  docker_password = var.docker_password
}

output "alb_url" {
  value = module.madabank.alb_url
}

output "alb_dns_name" {
  value = module.madabank.alb_dns_name
}