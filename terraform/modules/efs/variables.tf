# ==================================================
# EFS Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (production, staging, homolog)"
  type        = string

  validation {
    condition     = contains(["production", "staging", "homolog"], var.environment)
    error_message = "Environment must be production, staging, or homolog."
  }
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for EFS mount targets"
  type        = list(string)

  validation {
    condition     = length(var.private_subnet_ids) >= 2
    error_message = "At least 2 subnets required for EFS high availability."
  }
}

variable "efs_security_group_id" {
  description = "Security group ID for EFS mount targets"
  type        = string
}

variable "performance_mode" {
  description = "EFS performance mode (generalPurpose or maxIO)"
  type        = string
  default     = "generalPurpose"

  validation {
    condition     = contains(["generalPurpose", "maxIO"], var.performance_mode)
    error_message = "Performance mode must be generalPurpose or maxIO."
  }
}

variable "throughput_mode" {
  description = "EFS throughput mode (bursting or provisioned)"
  type        = string
  default     = "bursting"

  validation {
    condition     = contains(["bursting", "provisioned"], var.throughput_mode)
    error_message = "Throughput mode must be bursting or provisioned."
  }
}

variable "transition_to_ia" {
  description = "Transition files to Infrequent Access storage class (AFTER_7_DAYS, AFTER_14_DAYS, AFTER_30_DAYS, AFTER_60_DAYS, AFTER_90_DAYS)"
  type        = string
  default     = "AFTER_30_DAYS"

  validation {
    condition     = contains(["AFTER_7_DAYS", "AFTER_14_DAYS", "AFTER_30_DAYS", "AFTER_60_DAYS", "AFTER_90_DAYS"], var.transition_to_ia)
    error_message = "Invalid transition period."
  }
}

variable "enable_backup" {
  description = "Enable automatic EFS backups (via AWS Backup)"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
