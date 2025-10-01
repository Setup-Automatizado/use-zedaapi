# ==================================================
# Homolog Environment - Outputs
# ==================================================

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "alb_dns_name" {
  description = "ALB DNS name (use for DNS records)"
  value       = module.alb.alb_dns_name
}

output "alb_zone_id" {
  description = "ALB zone ID (for Route53 alias records)"
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
  description = "ECS Task Definition ARN (for CD workflow)"
  value       = module.ecs_service.task_definition_arn
}

output "secrets_arn" {
  description = "Secrets Manager ARN"
  value       = module.secrets.secret_arn
}

output "efs_file_system_id" {
  description = "EFS File System ID"
  value       = module.efs.efs_file_system_id
}
