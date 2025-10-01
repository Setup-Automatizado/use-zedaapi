# ==================================================
# ALB Module - Application Load Balancer
# ==================================================
# Creates ALB with target groups for:
# - API (port 8080)
# - MinIO Console (port 9001, optional)
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
# Application Load Balancer
# ==================================================
resource "aws_lb" "main" {
  name               = "${var.environment}-whatsmeow-alb"
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
      Name = "${var.environment}-whatsmeow-alb"
    }
  )
}

# ==================================================
# Target Group: API
# ==================================================
resource "aws_lb_target_group" "api" {
  name_prefix = "${substr(var.environment, 0, 3)}-api-"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    path                = "/healthz"
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
      Name = "${var.environment}-whatsmeow-api-tg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# ==================================================
# Target Group: MinIO Console (optional)
# ==================================================
resource "aws_lb_target_group" "minio" {
  count = var.expose_minio_console ? 1 : 0

  name_prefix = "${substr(var.environment, 0, 3)}-mio-"
  port        = 9001
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    path                = "/minio/health/live"
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
      Name = "${var.environment}-whatsmeow-minio-tg"
    }
  )

  lifecycle {
    create_before_destroy = true
  }
}

# ==================================================
# Listener: HTTP (redirect to HTTPS if certificate provided)
# ==================================================
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
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

    # Forward to API if no certificate
    target_group_arn = var.certificate_arn != null ? null : aws_lb_target_group.api.arn
  }

  tags = var.tags
}

# ==================================================
# Listener: HTTPS (if certificate provided)
# ==================================================
resource "aws_lb_listener" "https" {
  count = var.certificate_arn != null ? 1 : 0

  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }

  tags = var.tags
}

# ==================================================
# Listener Rule: MinIO Console (if enabled)
# ==================================================
resource "aws_lb_listener_rule" "minio" {
  count = var.expose_minio_console && var.certificate_arn != null ? 1 : 0

  listener_arn = aws_lb_listener.https[0].arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.minio[0].arn
  }

  condition {
    path_pattern {
      values = ["/minio/*"]
    }
  }

  tags = var.tags
}

# ==================================================
# Listener Rule: MinIO Console (HTTP fallback)
# ==================================================
resource "aws_lb_listener_rule" "minio_http" {
  count = var.expose_minio_console && var.certificate_arn == null ? 1 : 0

  listener_arn = aws_lb_listener.http.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.minio[0].arn
  }

  condition {
    path_pattern {
      values = ["/minio/*"]
    }
  }

  tags = var.tags
}
