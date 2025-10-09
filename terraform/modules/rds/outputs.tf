# ==================================================
# RDS Module - Outputs
# ==================================================

output "db_instance_identifier" {
  description = "RDS instance identifier"
  value       = aws_db_instance.this.id
}

output "db_instance_arn" {
  description = "RDS instance ARN"
  value       = aws_db_instance.this.arn
}

output "db_endpoint" {
  description = "RDS endpoint address"
  value       = aws_db_instance.this.address
}

output "db_port" {
  description = "RDS endpoint port"
  value       = aws_db_instance.this.port
}

output "db_subnet_group_name" {
  description = "Subnet group name"
  value       = aws_db_subnet_group.this.name
}

