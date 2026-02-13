# ==================================================
# ALB Manager Module - Outputs
# ==================================================

output "alb_arn" {
  description = "ARN of the Manager ALB"
  value       = aws_lb.manager.arn
}

output "alb_dns_name" {
  description = "DNS name of the Manager ALB"
  value       = aws_lb.manager.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the Manager ALB"
  value       = aws_lb.manager.zone_id
}

output "target_group_arn" {
  description = "ARN of the Manager target group"
  value       = aws_lb_target_group.manager.arn
}

output "http_listener_arn" {
  description = "ARN of the HTTP listener"
  value       = aws_lb_listener.http.arn
}

output "https_listener_arn" {
  description = "ARN of the HTTPS listener (if enabled)"
  value       = var.certificate_arn != null ? aws_lb_listener.https[0].arn : null
}

output "manager_url" {
  description = "URL to access the Manager"
  value       = var.certificate_arn != null ? "https://${aws_lb.manager.dns_name}" : "http://${aws_lb.manager.dns_name}"
}
