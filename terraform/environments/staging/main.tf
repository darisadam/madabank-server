terraform {
  backend "s3" {
    bucket         = "madabank-terraform-state-staging"
    key            = "staging/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "madabank-terraform-locks"
  }
}

module "madabank" {
  source = "../../"

  aws_region  = "us-east-1"
  environment = "staging"
  owner       = "darisadam.dev@gmail.com"

  # Networking
  vpc_cidr             = "10.1.0.0/16"
  public_subnet_cidrs  = ["10.1.1.0/24", "10.1.2.0/24"]
  private_subnet_cidrs = ["10.1.11.0/24", "10.1.12.0/24"]

  # Database - medium instance for staging
  db_instance_class       = "db.t3.small"
  db_allocated_storage    = 50
  backup_retention_period = 7
  db_multi_az            = false

  # Redis
  redis_node_type      = "cache.t3.small"
  redis_num_cache_nodes = 1

  # ECS - moderate resources
  container_cpu    = 512
  container_memory = 1024
  desired_count    = 2

  # Monitoring
  alert_email = "darisadam.dev@gmail.com"
}

output "alb_url" {
  value = module.madabank.alb_url
}