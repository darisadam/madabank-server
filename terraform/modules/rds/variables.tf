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

variable "db_security_group_id" {
  description = "DB security group ID"
  type        = string
}

variable "db_instance_class" {
  description = "DB instance class"
  type        = string
}

variable "db_allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_username" {
  description = "Database username"
  type        = string
}

variable "backup_retention_period" {
  description = "Backup retention period in days"
  type        = number
}

variable "multi_az" {
  description = "Enable multi-AZ"
  type        = bool
}