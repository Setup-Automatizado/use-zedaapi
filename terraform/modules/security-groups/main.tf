# ==================================================
# Security Groups Module
# ==================================================
# Creates security groups for:
# - ALB (public internet access)
# - ECS Tasks (internal communication)
# - RDS PostgreSQL
# - ElastiCache Redis
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
# Security Group: RDS PostgreSQL
# ==================================================
resource "aws_security_group" "rds" {
  name_prefix = "${var.environment}-whatsmeow-rds-"
  description = "Security group for PostgreSQL RDS instance"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-rds-sg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc_security_group_ingress_rule" "rds_from_ecs" {
  security_group_id = aws_security_group.rds.id
  description       = "Allow PostgreSQL from ECS tasks"

  referenced_security_group_id = aws_security_group.ecs_tasks.id
  from_port                    = var.rds_port
  to_port                      = var.rds_port
  ip_protocol                  = "tcp"

  tags = {
    Name = "postgres-from-ecs"
  }
}

# Allow RDS access from admin IPs (for database management)
resource "aws_vpc_security_group_ingress_rule" "rds_from_admin" {
  count = length(var.allowed_admin_ips)

  security_group_id = aws_security_group.rds.id
  description       = "Allow PostgreSQL from admin IP ${var.allowed_admin_ips[count.index]}"

  # If IP already has CIDR notation (contains '/'), use as-is, otherwise add /32
  cidr_ipv4   = can(regex("/", var.allowed_admin_ips[count.index])) ? var.allowed_admin_ips[count.index] : "${var.allowed_admin_ips[count.index]}/32"
  from_port   = var.rds_port
  to_port     = var.rds_port
  ip_protocol = "tcp"

  tags = {
    Name = "postgres-from-admin-${count.index}"
  }
}

resource "aws_vpc_security_group_egress_rule" "rds_all" {
  security_group_id = aws_security_group.rds.id
  description       = "Allow outbound traffic"

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"

  tags = {
    Name = "rds-egress"
  }
}

# ==================================================
# Security Group: ElastiCache Redis
# ==================================================
resource "aws_security_group" "redis" {
  name_prefix = "${var.environment}-whatsmeow-redis-"
  description = "Security group for ElastiCache Redis"
  vpc_id      = var.vpc_id

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-redis-sg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc_security_group_ingress_rule" "redis_from_ecs" {
  security_group_id = aws_security_group.redis.id
  description       = "Allow Redis from ECS tasks"

  referenced_security_group_id = aws_security_group.ecs_tasks.id
  from_port                    = var.redis_port
  to_port                      = var.redis_port
  ip_protocol                  = "tcp"

  tags = {
    Name = "redis-from-ecs"
  }
}

resource "aws_vpc_security_group_egress_rule" "redis_all" {
  security_group_id = aws_security_group.redis.id
  description       = "Allow outbound traffic"

  cidr_ipv4   = "0.0.0.0/0"
  ip_protocol = "-1"

  tags = {
    Name = "redis-egress"
  }
}

# Outbound: Not needed for EFS (mount targets are ingress-only)
