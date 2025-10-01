# ==================================================
# ECS Cluster Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (production, staging, homolog)"
  type        = string

  validation {
    condition     = contains(["production", "staging", "homolog"], var.environment)
    error_message = "Environment must be production, staging, or homolog."
  }
}

variable "enable_container_insights" {
  description = "Enable CloudWatch Container Insights (adds cost)"
  type        = bool
  default     = false
}

variable "enable_fargate_spot" {
  description = "Enable FARGATE_SPOT capacity provider (70% cost savings, may be interrupted)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
