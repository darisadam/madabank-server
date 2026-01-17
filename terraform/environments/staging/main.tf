terraform {
  backend "s3" {
    bucket         = "madabank-terraform-state-dev"
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

  # Networking - Different CIDR to allow potential peering later
  vpc_cidr             = "10.1.0.0/16"
  public_subnet_cidrs  = ["10.1.1.0/24", "10.1.2.0/24"]
  private_subnet_cidrs = ["10.1.11.0/24", "10.1.12.0/24"]

  # Database
  db_name             = "madabank_staging"
  db_username         = "madabank_admin"
  db_instance_class   = "db.t3.micro"
  db_allocated_storage = 20
  backup_retention_period = 7
  db_multi_az         = false

  # Redis
  redis_node_type       = "cache.t3.micro"
  redis_num_cache_nodes = 1

  # ECS
  container_cpu    = 256
  container_memory = 512
  desired_count    = 1

  # Monitoring
  alert_email = "darisadam.dev@gmail.com"

  # Container image
  container_image = "ghcr.io/darisadam/madabank-server:latest"
}

output "alb_url" {
  value = module.madabank.alb_url
}