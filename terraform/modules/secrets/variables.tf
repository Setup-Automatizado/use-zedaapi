# ==================================================
# Secrets Manager Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (production, staging, homolog)"
  type        = string

  validation {
    condition     = contains(["production", "staging", "homolog"], var.environment)
    error_message = "Environment must be production, staging, or homolog."
  }
}

variable "db_user" {
  description = "Database username"
  type        = string
  default     = "whatsmeow"
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.db_password) >= 16
    error_message = "Database password must be at least 16 characters long."
  }
}

variable "minio_access_key" {
  description = "MinIO access key (username)"
  type        = string
  default     = "minio"
  sensitive   = true
}

variable "minio_secret_key" {
  description = "MinIO secret key (password)"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.minio_secret_key) >= 16
    error_message = "MinIO secret key must be at least 16 characters long."
  }
}

variable "recovery_window_in_days" {
  description = "Number of days to wait before permanently deleting the secret (0 = immediate deletion)"
  type        = number
  default     = 7

  validation {
    condition     = var.recovery_window_in_days == 0 || (var.recovery_window_in_days >= 7 && var.recovery_window_in_days <= 30)
    error_message = "Recovery window must be 0 (immediate) or between 7 and 30 days."
  }
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
