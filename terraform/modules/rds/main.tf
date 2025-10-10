# ==================================================
# RDS Module - PostgreSQL Instance
# ==================================================

terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

locals {
  identifier = "${var.environment}-whatsmeow-db"
}

# ==================================================
# Subnet Group
# ==================================================
resource "aws_db_subnet_group" "this" {
  name       = local.identifier
  subnet_ids = var.subnet_ids

  tags = merge(
    var.tags,
    {
      Name = "${local.identifier}-subnet-group"
    }
  )
}

# ==================================================
# PostgreSQL Instance
# ==================================================
resource "aws_db_instance" "this" {
  identifier                            = local.identifier
  engine                                = "postgres"
  engine_version                        = var.engine_version
  instance_class                        = var.instance_class
  db_name                               = var.db_name
  username                              = var.db_username
  password                              = var.db_password
  allocated_storage                     = var.allocated_storage
  max_allocated_storage                 = var.max_allocated_storage
  multi_az                              = var.multi_az
  storage_encrypted                     = true
  auto_minor_version_upgrade            = true
  backup_retention_period               = var.backup_retention_period
  deletion_protection                   = var.deletion_protection
  skip_final_snapshot                   = var.skip_final_snapshot
  final_snapshot_identifier             = var.skip_final_snapshot ? null : "${local.identifier}-final-${random_id.final_snapshot[0].hex}"
  apply_immediately                     = var.apply_immediately
  db_subnet_group_name                  = aws_db_subnet_group.this.name
  vpc_security_group_ids                = var.security_group_ids
  publicly_accessible                   = false
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention : null
  copy_tags_to_snapshot                 = true
  monitoring_interval                   = var.monitoring_interval
  monitoring_role_arn                   = var.monitoring_interval > 0 ? var.monitoring_role_arn : null

  tags = merge(
    var.tags,
    {
      Name = local.identifier
    }
  )
}

resource "random_id" "final_snapshot" {
  count       = var.skip_final_snapshot ? 0 : 1
  byte_length = 4
}
