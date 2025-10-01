# ==================================================
# ALB Module - Outputs
# ==================================================

output "alb_id" {
  description = "Application Load Balancer ID"
  value       = aws_lb.main.id
}

output "alb_arn" {
  description = "Application Load Balancer ARN"
  value       = aws_lb.main.arn
}

output "alb_dns_name" {
  description = "Application Load Balancer DNS name"
  value       = aws_lb.main.dns_name
}

output "alb_zone_id" {
  description = "Application Load Balancer zone ID (for Route53)"
  value       = aws_lb.main.zone_id
}

output "api_target_group_arn" {
  description = "API target group ARN"
  value       = aws_lb_target_group.api.arn
}

output "api_target_group_name" {
  description = "API target group name"
  value       = aws_lb_target_group.api.name
}

output "minio_target_group_arn" {
  description = "MinIO target group ARN (empty if not exposed)"
  value       = var.expose_minio_console ? aws_lb_target_group.minio[0].arn : null
}

output "http_listener_arn" {
  description = "HTTP listener ARN"
  value       = aws_lb_listener.http.arn
}

output "https_listener_arn" {
  description = "HTTPS listener ARN (null if no certificate)"
  value       = var.certificate_arn != null ? aws_lb_listener.https[0].arn : null
}
