# ==================================================
# ECS Service Module - Outputs
# ==================================================

output "task_definition_arn" {
  description = "ECS Task Definition ARN"
  value       = aws_ecs_task_definition.main.arn
}

output "task_definition_family" {
  description = "ECS Task Definition family"
  value       = aws_ecs_task_definition.main.family
}

output "task_definition_revision" {
  description = "ECS Task Definition revision"
  value       = aws_ecs_task_definition.main.revision
}

output "service_name" {
  description = "ECS Service name"
  value       = aws_ecs_service.main.name
}

output "service_id" {
  description = "ECS Service ID"
  value       = aws_ecs_service.main.id
}

output "task_execution_role_arn" {
  description = "IAM Role ARN for task execution"
  value       = aws_iam_role.task_execution.arn
}

output "task_role_arn" {
  description = "IAM Role ARN for task"
  value       = aws_iam_role.task.arn
}
