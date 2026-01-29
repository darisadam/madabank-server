variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "madabank"
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "owner" {
  description = "Project owner email"
  type        = string
  default     = ""
}

# Networking
variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "Public subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "Private subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24"]
}

variable "single_nat_gateway" {
  description = "Use a single NAT Gateway (saves EIPs)"
  type        = bool
  default     = false
}

# Database
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "db_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 20
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "madabank"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "madabankadmin"
}

variable "backup_retention_period" {
  description = "Database backup retention period in days"
  type        = number
  default     = 7
}

variable "db_multi_az" {
  description = "Enable multi-AZ for RDS"
  type        = bool
  default     = false
}

# Redis
variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_cache_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 1
}

# ECS
variable "container_image" {
  description = "Docker container image"
  type        = string
  default     = "ghcr.io/darisadam/madabank-server:latest"
}

variable "container_cpu" {
  description = "Container CPU units"
  type        = number
  default     = 256
}

variable "container_memory" {
  description = "Container memory in MB"
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "Desired number of ECS tasks"
  type        = number
  default     = 2
}

variable "min_capacity" {
  description = "Minimum task count for autoscaling"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Maximum task count for autoscaling"
  type        = number
  default     = 5
}

# ALB
variable "certificate_arn" {
  description = "ACM certificate ARN for HTTPS"
  type        = string
}

# Monitoring
variable "alert_email" {
  description = "Email for CloudWatch alerts"
  type        = string
}

# Docker Registry Credentials
variable "docker_username" {
  description = "Username for Docker Registry"
  type        = string
  sensitive   = true
}

variable "docker_password" {
  description = "Password/Token for Docker Registry"
  type        = string
  sensitive   = true
}
