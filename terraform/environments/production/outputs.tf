# ==================================================
# Production Environment - Outputs
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
# ECR Outputs
# ==================================================

output "ecr_api_repository_url" {
  description = "ECR repository URL for API"
  value       = module.ecr.api_repository_url
}

output "ecr_manager_repository_url" {
  description = "ECR repository URL for Manager"
  value       = module.ecr.manager_repository_url
}

# ==================================================
# CloudFront Outputs
# ==================================================

output "api_https_url" {
  description = "API HTTPS URL via CloudFront"
  value       = module.cloudfront_api.https_url
}

output "api_cloudfront_distribution_id" {
  description = "API CloudFront distribution ID"
  value       = module.cloudfront_api.distribution_id
}

# ==================================================
# Manager Outputs
# ==================================================

output "manager_alb_dns_name" {
  description = "Manager ALB DNS name"
  value       = var.enable_manager ? module.alb_manager[0].alb_dns_name : null
}

output "manager_https_url" {
  description = "Manager HTTPS URL via CloudFront"
  value       = var.enable_manager ? module.cloudfront_manager[0].https_url : null
}

output "manager_cloudfront_distribution_id" {
  description = "Manager CloudFront distribution ID"
  value       = var.enable_manager ? module.cloudfront_manager[0].distribution_id : null
}

output "manager_service_name" {
  description = "Manager ECS service name"
  value       = var.enable_manager ? module.ecs_service_manager[0].service_name : null
}
