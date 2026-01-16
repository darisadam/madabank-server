variable "project_name" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "ECS security group ID"
  type        = string
}

variable "ecs_task_execution_role_arn" {
  description = "ECS task execution role ARN"
  type        = string
}

variable "ecs_task_role_arn" {
  description = "ECS task role ARN"
  type        = string
}

variable "alb_target_group_arn" {
  description = "ALB target group ARN"
  type        = string
}

variable "db_endpoint" {
  description = "Database endpoint"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_username" {
  description = "Database username"
  type        = string
}

variable "db_password_secret_arn" {
  description = "Database password secret ARN"
  type        = string
}

variable "redis_endpoint" {
  description = "Redis endpoint"
  type        = string
}

variable "container_image" {
  description = "Container image"
  type        = string
}

variable "container_cpu" {
  description = "Container CPU"
  type        = number
}

variable "container_memory" {
  description = "Container memory"
  type        = number
}

variable "desired_count" {
  description = "Desired task count"
  type        = number
}

variable "min_capacity" {
  description = "Minimum task count for autoscaling"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Maximum task count for autoscaling"
  type        = number
  default     = 10
}

variable "jwt_secret_arn" {
  description = "JWT secret ARN"
  type        = string
}

variable "encryption_key_arn" {
  description = "Encryption key ARN"
  type        = string
}