terraform {
  backend "s3" {
    bucket         = "madabank-terraform-state-prod"
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

  # Database - production-grade with multi-AZ
  db_instance_class       = "db.t3.medium"
  db_allocated_storage    = 100
  backup_retention_period = 30
  db_multi_az            = true

  # Redis - production with failover
  redis_node_type      = "cache.t3.medium"
  redis_num_cache_nodes = 2

  # ECS - production resources with autoscaling
  container_cpu    = 1024
  container_memory = 2048
  desired_count    = 3

  # Monitoring
  alert_email = "alerts@madabank.com"
}

output "alb_url" {
  value = module.madabank.alb_url
}