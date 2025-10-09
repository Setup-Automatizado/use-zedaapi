# ==================================================
# Secrets Manager Module
# ==================================================
# Creates AWS Secrets Manager secret with caller-provided payload
# Auto-rotation is disabled by default (use Secrets Manager rotation
# or external workflows for rotation policies)
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
  secret_id     = aws_secretsmanager_secret.main.id
  secret_string = jsonencode(var.secret_payload)
}
