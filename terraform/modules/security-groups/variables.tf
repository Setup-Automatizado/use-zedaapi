# ==================================================
# Security Groups Module - Variables
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
  description = "VPC ID where security groups will be created"
  type        = string
}

variable "expose_minio_console" {
  description = "Expose MinIO console through ALB (development/debugging only)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
