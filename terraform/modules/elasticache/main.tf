# ==================================================
# ElastiCache Module - Redis Replication Group
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

locals {
  replication_group_id = "${var.environment}-whatsmeow-redis"
}

# ==================================================
# Subnet Group
# ==================================================
resource "aws_elasticache_subnet_group" "this" {
  name       = local.replication_group_id
  subnet_ids = var.subnet_ids

  tags = merge(
    var.tags,
    {
      Name = "${local.replication_group_id}-subnet-group"
    }
  )
}

# ==================================================
# Redis Replication Group
# ==================================================
resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = local.replication_group_id
  description                = "Redis for WhatsApp API (${var.environment})"
  engine                     = "redis"
  engine_version             = var.engine_version
  node_type                  = var.node_type
  parameter_group_name       = "default.redis7"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.this.name
  security_group_ids         = var.security_group_ids
  automatic_failover_enabled = var.replicas_per_node_group > 0 ? var.automatic_failover_enabled : false
  multi_az_enabled           = var.replicas_per_node_group > 0 ? var.multi_az_enabled : false
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  num_node_groups            = 1
  replicas_per_node_group    = var.replicas_per_node_group
  apply_immediately          = true
  auto_minor_version_upgrade = true
  maintenance_window         = "sun:06:00-sun:07:00"
  snapshot_window            = "03:00-04:00"
  snapshot_retention_limit   = 3
  auth_token                 = length(var.auth_token) > 0 ? var.auth_token : null
  auth_token_update_strategy = length(var.auth_token) > 0 ? "ROTATE" : null

  tags = merge(
    var.tags,
    {
      Name = local.replication_group_id
    }
  )
}
