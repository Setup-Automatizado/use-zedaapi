# ==================================================
# CloudFront Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (e.g., production, staging, homolog)"
  type        = string
}

variable "service_name" {
  description = "Service name for identification (e.g., api, manager)"
  type        = string
}

variable "alb_dns_name" {
  description = "DNS name of the ALB to use as origin"
  type        = string
}

variable "origin_protocol_policy" {
  description = "Protocol policy for origin (http-only or https-only)"
  type        = string
  default     = "http-only"
}

variable "origin_read_timeout" {
  description = "Origin read timeout in seconds"
  type        = number
  default     = 60
}

variable "price_class" {
  description = "CloudFront price class"
  type        = string
  default     = "PriceClass_100"
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
