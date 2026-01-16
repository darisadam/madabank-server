# ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "main" {
  name_prefix = "${var.project_name}-${var.environment}-"
  subnet_ids  = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-${var.environment}-redis-subnet-group"
  }
}

# ElastiCache Parameter Group
resource "aws_elasticache_parameter_group" "main" {
  name_prefix = "${var.project_name}-${var.environment}-"
  family      = "redis7"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-redis-parameter-group"
  }

  lifecycle {
    create_before_destroy = true
  }
}

# ElastiCache Replication Group (Redis)
resource "aws_elasticache_replication_group" "main" {
  replication_group_id       = "${var.project_name}-${var.environment}"
  replication_group_description = "Redis cluster for ${var.project_name} ${var.environment}"
  
  engine               = "redis"
  engine_version       = "7.0"
  node_type            = var.redis_node_type
  num_cache_clusters   = var.redis_num_cache_nodes
  port                 = 6379
  
  parameter_group_name = aws_elasticache_parameter_group.main.name
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [var.redis_security_group_id]
  
  automatic_failover_enabled = var.redis_num_cache_nodes > 1
  multi_az_enabled          = var.redis_num_cache_nodes > 1
  
  at_rest_encryption_enabled = true
  transit_encryption_enabled = false  # Set to true if you want in-transit encryption
  
  snapshot_retention_limit = 5
  snapshot_window         = "03:00-05:00"
  maintenance_window      = "mon:05:00-mon:07:00"

  auto_minor_version_upgrade = true
  
  tags = {
    Name = "${var.project_name}-${var.environment}-redis"
  }
}