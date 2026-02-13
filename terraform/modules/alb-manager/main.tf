# ==================================================
# ALB Manager Module - Application Load Balancer
# ==================================================
# Creates ALB dedicado para o Manager Frontend (Next.js)
# Separado do ALB da API para evitar conflitos de rotas
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
# Application Load Balancer - Manager
# ==================================================
resource "aws_lb" "manager" {
  name               = "${var.environment}-manager-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.alb_security_group_id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection       = var.enable_deletion_protection
  enable_cross_zone_load_balancing = true
  enable_http2                     = true

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-manager-alb"
    }
  )
}

# ==================================================
# Target Group: Manager Frontend (Next.js)
# ==================================================
resource "aws_lb_target_group" "manager" {
  name_prefix = substr("${var.environment}mgr", 0, 6)
  port        = 3000
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    path                = "/api/health"
    protocol            = "HTTP"
    matcher             = "200"
    interval            = 30
    timeout             = 10
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  deregistration_delay = 30

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-manager-tg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# ==================================================
# Listener: HTTP
# ==================================================
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.manager.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = var.certificate_arn != null ? "redirect" : "forward"

    # Redirect to HTTPS if certificate available
    dynamic "redirect" {
      for_each = var.certificate_arn != null ? [1] : []
      content {
        port        = "443"
        protocol    = "HTTPS"
        status_code = "HTTP_301"
      }
    }

    # Forward to Manager
    target_group_arn = var.certificate_arn != null ? null : aws_lb_target_group.manager.arn
  }

  tags = var.tags
}

# ==================================================
# Listener: HTTPS (if certificate provided)
# ==================================================
resource "aws_lb_listener" "https" {
  count = var.certificate_arn != null ? 1 : 0

  load_balancer_arn = aws_lb.manager.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.manager.arn
  }

  tags = var.tags
}
