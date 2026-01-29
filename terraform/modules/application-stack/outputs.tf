output "vpc_id" {
  value = module.networking.vpc_id
}

output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "alb_url" {
  value = module.alb.alb_url
}

output "db_endpoint" {
  value = module.rds.db_endpoint
}

output "redis_endpoint" {
  value = module.elasticache.redis_endpoint
}

output "dashboard_url" {
  value = module.monitoring.dashboard_url
}
