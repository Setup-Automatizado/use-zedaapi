# ==================================================
# ElastiCache Module - Outputs
# ==================================================

output "primary_endpoint" {
  description = "Primary Redis endpoint"
  value       = aws_elasticache_replication_group.this.primary_endpoint_address
}

output "reader_endpoint" {
  description = "Reader Redis endpoint"
  value       = aws_elasticache_replication_group.this.reader_endpoint_address
}

output "port" {
  description = "Redis port"
  value       = aws_elasticache_replication_group.this.port
}

output "replication_group_id" {
  description = "Replication group identifier"
  value       = aws_elasticache_replication_group.this.id
}

