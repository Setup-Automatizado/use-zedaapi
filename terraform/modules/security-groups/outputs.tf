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

output "efs_security_group_id" {
  description = "Security Group ID for EFS mount targets"
  value       = aws_security_group.efs.id
}
