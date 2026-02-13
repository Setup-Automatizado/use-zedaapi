# ==================================================
# Homolog Environment - Outputs
# ==================================================

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "alb_dns_name" {
  description = "ALB DNS name"
  value       = module.alb.alb_dns_name
}

output "alb_zone_id" {
  description = "ALB zone ID"
  value       = module.alb.alb_zone_id
}

output "ecs_cluster_name" {
  description = "ECS Cluster name"
  value       = module.ecs_cluster.cluster_name
}

output "ecs_service_name" {
  description = "ECS Service name"
  value       = module.ecs_service.service_name
}

output "task_definition_arn" {
  description = "ECS Task Definition ARN"
  value       = module.ecs_service.task_definition_arn
}

output "secrets_arn" {
  description = "Secrets Manager ARN"
  value       = module.secrets.secret_arn
}

output "rds_endpoint" {
  description = "PostgreSQL endpoint"
  value       = module.rds.db_endpoint
}

output "redis_endpoint" {
  description = "Redis primary endpoint"
  value       = module.elasticache.primary_endpoint
}

output "s3_bucket_name" {
  description = "Media bucket"
  value       = module.s3.bucket_name
}

# ==================================================
# Manager Outputs
# ==================================================

output "manager_alb_dns_name" {
  description = "Manager ALB DNS name"
  value       = var.enable_manager ? module.alb_manager[0].alb_dns_name : null
}

output "manager_url" {
  description = "Manager URL"
  value       = var.enable_manager ? module.alb_manager[0].manager_url : null
}

output "api_url" {
  description = "API URL (ALB da API)"
  value       = "http://${module.alb.alb_dns_name}"
}
