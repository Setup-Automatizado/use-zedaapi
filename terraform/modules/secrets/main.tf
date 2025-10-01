# ==================================================
# Secrets Manager Module
# ==================================================
# Creates AWS Secrets Manager secret with:
# - Database credentials
# - MinIO credentials
# - Auto-rotation disabled (manual rotation recommended)
# ==================================================

terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ==================================================
# Secrets Manager Secret
# ==================================================
resource "aws_secretsmanager_secret" "main" {
  name        = "${var.environment}/whatsmeow/credentials"
  description = "Credentials for WhatsApp API (${var.environment})"

  recovery_window_in_days = var.recovery_window_in_days

  tags = merge(
    var.tags,
    {
      Name        = "${var.environment}-whatsmeow-secrets"
      Environment = var.environment
    }
  )
}

# ==================================================
# Secret Version (Initial Values)
# ==================================================
resource "aws_secretsmanager_secret_version" "main" {
  secret_id = aws_secretsmanager_secret.main.id

  secret_string = jsonencode({
    db_user          = var.db_user
    db_password      = var.db_password
    minio_access_key = var.minio_access_key
    minio_secret_key = var.minio_secret_key
  })
}
