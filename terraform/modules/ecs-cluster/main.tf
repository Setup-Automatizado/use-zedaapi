# ==================================================
# ECS Cluster Module
# ==================================================
# Creates ECS Fargate cluster with:
# - CloudWatch Container Insights (optional)
# - Capacity providers (FARGATE and FARGATE_SPOT)
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
# ECS Cluster
# ==================================================
resource "aws_ecs_cluster" "main" {
  name = "${var.environment}-whatsmeow-cluster"

  setting {
    name  = "containerInsights"
    value = var.enable_container_insights ? "enabled" : "disabled"
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-cluster"
    }
  )
}

# ==================================================
# Cluster Capacity Providers
# ==================================================
resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = var.enable_fargate_spot ? ["FARGATE", "FARGATE_SPOT"] : ["FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
    base              = 1
  }

  # Use FARGATE_SPOT for additional tasks (optional)
  dynamic "default_capacity_provider_strategy" {
    for_each = var.enable_fargate_spot ? [1] : []
    content {
      capacity_provider = "FARGATE_SPOT"
      weight            = 4 # 80% FARGATE_SPOT, 20% FARGATE
      base              = 0
    }
  }
}
