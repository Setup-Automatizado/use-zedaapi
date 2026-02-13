# ==================================================
# Secrets Manager Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (production, staging, homolog) with optional suffix"
  type        = string

  validation {
    condition     = can(regex("^(production|staging|homolog)(-[a-z]+)?$", var.environment))
    error_message = "Environment must be production, staging, or homolog (with optional suffix like -manager)."
  }
}

variable "secret_payload" {
  description = "Key/value pairs stored in the secret"
  type        = map(string)
  sensitive   = true
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
