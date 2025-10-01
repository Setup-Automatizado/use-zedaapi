# ==================================================
# Security Groups Module
# ==================================================
# Creates security groups for:
# - ALB (public internet access)
# - ECS Tasks (internal communication)
# - EFS (file system access)
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
# Security Group: Application Load Balancer
# ==================================================
resource "aws_security_group" "alb" {
  name_prefix = "${var.environment}-whatsmeow-alb-"
  description = "Security group for Application Load Balancer"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-alb-sg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# Inbound: HTTP from anywhere
resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  description       = "Allow HTTP from internet"

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 80
  to_port     = 80
  ip_protocol = "tcp"

  tags = {
    Name = "allow-http"
  }
}

# Inbound: HTTPS from anywhere
resource "aws_vpc_security_group_ingress_rule" "alb_https" {
  security_group_id = aws_security_group.alb.id
  description       = "Allow HTTPS from internet"

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 443
  to_port     = 443
  ip_protocol = "tcp"

  tags = {
    Name = "allow-https"
  }
}

# Outbound: All traffic to ECS tasks
resource "aws_vpc_security_group_egress_rule" "alb_to_ecs" {
  security_group_id = aws_security_group.alb.id
  description       = "Allow all traffic to ECS tasks"

  referenced_security_group_id = aws_security_group.ecs_tasks.id
  ip_protocol                  = "-1"

  tags = {
    Name = "alb-to-ecs"
  }
}

# ==================================================
# Security Group: ECS Tasks
# ==================================================
resource "aws_security_group" "ecs_tasks" {
  name_prefix = "${var.environment}-whatsmeow-ecs-"
  description = "Security group for ECS tasks"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-ecs-sg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# Inbound: API port from ALB
resource "aws_vpc_security_group_ingress_rule" "ecs_api_from_alb" {
  security_group_id = aws_security_group.ecs_tasks.id
  description       = "Allow API traffic from ALB"

  referenced_security_group_id = aws_security_group.alb.id
  from_port                    = 8080
  to_port                      = 8080
  ip_protocol                  = "tcp"

  tags = {
    Name = "api-from-alb"
  }
}

# Inbound: MinIO console from ALB (optional, for debugging)
resource "aws_vpc_security_group_ingress_rule" "ecs_minio_from_alb" {
  count = var.expose_minio_console ? 1 : 0

  security_group_id = aws_security_group.ecs_tasks.id
  description       = "Allow MinIO console from ALB"

  referenced_security_group_id = aws_security_group.alb.id
  from_port                    = 9001
  to_port                      = 9001
  ip_protocol                  = "tcp"

  tags = {
    Name = "minio-console-from-alb"
  }
}

# Inbound: Self (for internal container communication)
resource "aws_vpc_security_group_ingress_rule" "ecs_self" {
  security_group_id = aws_security_group.ecs_tasks.id
  description       = "Allow all internal communication between containers"

  referenced_security_group_id = aws_security_group.ecs_tasks.id
  ip_protocol                  = "-1"

  tags = {
    Name = "ecs-internal"
  }
}

# Outbound: All traffic (for pulling images, external APIs)
resource "aws_vpc_security_group_egress_rule" "ecs_all" {
  security_group_id = aws_security_group.ecs_tasks.id
  description       = "Allow all outbound traffic"

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"

  tags = {
    Name = "ecs-outbound"
  }
}

# ==================================================
# Security Group: EFS
# ==================================================
resource "aws_security_group" "efs" {
  name_prefix = "${var.environment}-whatsmeow-efs-"
  description = "Security group for EFS mount targets"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-efs-sg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# Inbound: NFS from ECS tasks
resource "aws_vpc_security_group_ingress_rule" "efs_from_ecs" {
  security_group_id = aws_security_group.efs.id
  description       = "Allow NFS from ECS tasks"

  referenced_security_group_id = aws_security_group.ecs_tasks.id
  from_port                    = 2049
  to_port                      = 2049
  ip_protocol                  = "tcp"

  tags = {
    Name = "nfs-from-ecs"
  }
}

# Outbound: Not needed for EFS (mount targets are ingress-only)
