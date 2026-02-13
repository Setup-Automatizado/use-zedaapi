# ==================================================
# CloudFront Module - HTTPS Distribution
# ==================================================
# Provides HTTPS termination via CloudFront's default
# certificate (*.cloudfront.net) without requiring a
# custom domain or ACM certificate.
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
# CloudFront Distribution
# ==================================================
resource "aws_cloudfront_distribution" "this" {
  enabled         = true
  comment         = "${var.environment} - ${var.service_name}"
  is_ipv6_enabled = true
  price_class     = var.price_class
  http_version    = "http2and3"

  # Origin: ALB (HTTP only - CloudFront terminates TLS)
  origin {
    domain_name = var.alb_dns_name
    origin_id   = "${var.environment}-${var.service_name}-alb"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_protocol_policy   = var.origin_protocol_policy
      origin_ssl_protocols     = ["TLSv1.2"]
      origin_read_timeout      = var.origin_read_timeout
      origin_keepalive_timeout = 60
    }
  }

  # Default cache behavior - NO CACHING (dynamic content)
  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "${var.environment}-${var.service_name}-alb"

    # Use managed CachingDisabled policy
    cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingDisabled
    origin_request_policy_id = "216adef6-5c7f-47e4-b989-5492eafa07d3" # AllViewer

    viewer_protocol_policy = "redirect-to-https"
    compress               = true
  }

  # No geo restrictions
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # Use CloudFront default certificate (*.cloudfront.net)
  viewer_certificate {
    cloudfront_default_certificate = true
    minimum_protocol_version       = "TLSv1.2_2021"
  }

  tags = merge(var.tags, {
    Name    = "${var.environment}-${var.service_name}-cf"
    Service = var.service_name
  })
}
