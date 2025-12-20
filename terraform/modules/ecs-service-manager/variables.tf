# ==================================================
# ECS Service Manager Module - Variables
# ==================================================

variable "environment" {
  description = "Environment name (production, staging, homolog)"
  type        = string

  validation {
    condition     = contains(["production", "staging", "homolog"], var.environment)
    error_message = "Environment must be production, staging, or homolog."
  }
}

variable "cluster_id" {
  description = "ECS Cluster ID"
  type        = string
}

variable "cluster_name" {
  description = "ECS Cluster name"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for ECS tasks"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "Security group ID for ECS tasks"
  type        = string
}

variable "target_group_arn" {
  description = "ALB target group ARN for Manager"
  type        = string
}

variable "secrets_arn" {
  description = "Secrets Manager ARN containing credentials"
  type        = string
}

variable "aws_region" {
  description = "AWS region for runtime configuration"
  type        = string
}

# --------------------------------------------------
# Manager Application Configuration
# --------------------------------------------------

variable "manager_image" {
  description = "Docker image for Manager container"
  type        = string
}

variable "app_url" {
  description = "Public URL for the Manager application (NEXT_PUBLIC_APP_URL)"
  type        = string
}

variable "whatsapp_api_url" {
  description = "Internal URL for WhatsApp API backend"
  type        = string
}

# --------------------------------------------------
# S3 Configuration
# --------------------------------------------------

variable "s3_endpoint" {
  description = "S3/MinIO endpoint (empty for AWS S3)"
  type        = string
  default     = ""
}

variable "s3_bucket" {
  description = "S3 bucket name for media storage"
  type        = string
}

variable "s3_bucket_arn" {
  description = "S3 bucket ARN for IAM policies"
  type        = string
}

variable "s3_public_url" {
  description = "Public URL for S3 bucket"
  type        = string
  default     = ""
}

# --------------------------------------------------
# SMTP Configuration
# --------------------------------------------------

variable "smtp_host" {
  description = "SMTP server hostname"
  type        = string
  default     = ""
}

variable "smtp_port" {
  description = "SMTP server port"
  type        = number
  default     = 587
}

variable "smtp_user" {
  description = "SMTP username"
  type        = string
  default     = ""
}

# --------------------------------------------------
# Email Configuration
# --------------------------------------------------

variable "email_from_name" {
  description = "Email sender display name"
  type        = string
  default     = "WhatsApp Manager"
}

variable "email_from_address" {
  description = "Email sender address"
  type        = string
  default     = "noreply@example.com"
}

variable "support_email" {
  description = "Support email address"
  type        = string
  default     = "support@example.com"
}

# --------------------------------------------------
# Task Configuration
# --------------------------------------------------

variable "task_cpu" {
  description = "Task CPU units (256, 512, 1024, etc)"
  type        = number
  default     = 256

  validation {
    condition     = contains([256, 512, 1024, 2048, 4096], var.task_cpu)
    error_message = "Task CPU must be a valid Fargate value."
  }
}

variable "task_memory" {
  description = "Task memory in MB"
  type        = number
  default     = 512

  validation {
    condition     = var.task_memory >= 512 && var.task_memory <= 30720
    error_message = "Task memory must be between 512 MB and 30 GB."
  }
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 1

  validation {
    condition     = var.desired_count >= 0
    error_message = "Desired count must be >= 0."
  }
}

variable "enable_execute_command" {
  description = "Enable ECS Exec for debugging"
  type        = bool
  default     = false
}

variable "assign_public_ip" {
  description = "Assign public IP to ECS tasks"
  type        = bool
  default     = true
}

# --------------------------------------------------
# Auto Scaling Configuration
# --------------------------------------------------

variable "enable_autoscaling" {
  description = "Enable auto-scaling based on CPU/Memory"
  type        = bool
  default     = false
}

variable "autoscaling_min_capacity" {
  description = "Minimum number of tasks"
  type        = number
  default     = 1
}

variable "autoscaling_max_capacity" {
  description = "Maximum number of tasks"
  type        = number
  default     = 3
}

variable "autoscaling_cpu_target" {
  description = "Target CPU utilization percentage"
  type        = number
  default     = 70
}

variable "autoscaling_memory_target" {
  description = "Target memory utilization percentage"
  type        = number
  default     = 80
}

# --------------------------------------------------
# OAuth Configuration
# --------------------------------------------------

variable "github_client_id" {
  description = "GitHub OAuth Client ID"
  type        = string
  default     = ""
}

variable "google_client_id" {
  description = "Google OAuth Client ID"
  type        = string
  default     = ""
}

# --------------------------------------------------
# Tags
# --------------------------------------------------

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
