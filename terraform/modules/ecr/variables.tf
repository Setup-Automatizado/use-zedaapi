# ==================================================
# ECR Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (e.g., production, staging, homolog)"
  type        = string
}

variable "api_repository_name" {
  description = "ECR repository name for the API"
  type        = string
  default     = "whatsapp-api"
}

variable "manager_repository_name" {
  description = "ECR repository name for the Manager"
  type        = string
  default     = "manager-whatsapp-api"
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
