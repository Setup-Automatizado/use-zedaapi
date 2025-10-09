# ==================================================
# Security Groups Module - Outputs
# ==================================================

output "alb_security_group_id" {
  description = "Security Group ID for Application Load Balancer"
  value       = aws_security_group.alb.id
}

output "ecs_tasks_security_group_id" {
  description = "Security Group ID for ECS tasks"
  value       = aws_security_group.ecs_tasks.id
}

output "rds_security_group_id" {
  description = "Security Group ID for PostgreSQL RDS instance"
  value       = aws_security_group.rds.id
}

output "redis_security_group_id" {
  description = "Security Group ID for ElastiCache Redis"
  value       = aws_security_group.redis.id
}
