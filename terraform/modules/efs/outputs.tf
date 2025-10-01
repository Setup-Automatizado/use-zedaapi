# ==================================================
# EFS Module - Outputs
# ==================================================

output "efs_file_system_id" {
  description = "EFS file system ID"
  value       = aws_efs_file_system.main.id
}

output "efs_file_system_arn" {
  description = "EFS file system ARN"
  value       = aws_efs_file_system.main.arn
}

output "postgres_access_point_id" {
  description = "EFS Access Point ID for Postgres"
  value       = aws_efs_access_point.postgres.id
}

output "postgres_access_point_arn" {
  description = "EFS Access Point ARN for Postgres"
  value       = aws_efs_access_point.postgres.arn
}

output "redis_access_point_id" {
  description = "EFS Access Point ID for Redis"
  value       = aws_efs_access_point.redis.id
}

output "redis_access_point_arn" {
  description = "EFS Access Point ARN for Redis"
  value       = aws_efs_access_point.redis.arn
}

output "minio_access_point_id" {
  description = "EFS Access Point ID for MinIO"
  value       = aws_efs_access_point.minio.id
}

output "minio_access_point_arn" {
  description = "EFS Access Point ARN for MinIO"
  value       = aws_efs_access_point.minio.arn
}
