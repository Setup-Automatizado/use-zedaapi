# ==================================================
# ALB Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (production, staging, homolog)"
  type        = string

  validation {
    condition     = contains(["production", "staging", "homolog"], var.environment)
    error_message = "Environment must be production, staging, or homolog."
  }
}

variable "vpc_id" {
  description = "VPC ID where ALB will be created"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs for ALB"
  type        = list(string)

  validation {
    condition     = length(var.public_subnet_ids) >= 2
    error_message = "At least 2 public subnets required for ALB high availability."
  }
}

variable "alb_security_group_id" {
  description = "Security group ID for ALB"
  type        = string
}

variable "certificate_arn" {
  description = "ACM certificate ARN for HTTPS (optional, uses HTTP if not provided)"
  type        = string
  default     = null
}

variable "expose_minio_console" {
  description = "Expose MinIO console through ALB (development/debugging only)"
  type        = bool
  default     = false
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection for ALB (recommended for production)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
