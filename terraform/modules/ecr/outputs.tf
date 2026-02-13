# ==================================================
# ECR Module - Outputs
# ==================================================

output "api_repository_url" {
  description = "URL of the API ECR repository"
  value       = aws_ecr_repository.api.repository_url
}

output "api_repository_arn" {
  description = "ARN of the API ECR repository"
  value       = aws_ecr_repository.api.arn
}

output "manager_repository_url" {
  description = "URL of the Manager ECR repository"
  value       = aws_ecr_repository.manager.repository_url
}

output "manager_repository_arn" {
  description = "ARN of the Manager ECR repository"
  value       = aws_ecr_repository.manager.arn
}

output "registry_id" {
  description = "The registry ID where the repositories were created"
  value       = aws_ecr_repository.api.registry_id
}
