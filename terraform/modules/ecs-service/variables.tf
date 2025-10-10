# ==================================================
# ECS Service Module - Variables
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

variable "private_subnet_ids" {
  description = "List of private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "Security group ID for ECS tasks"
  type        = string
}

variable "api_target_group_arn" {
  description = "ALB target group ARN for API"
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

variable "db_host" {
  description = "Database endpoint hostname"
  type        = string
}

variable "db_port" {
  description = "Database port"
  type        = number
  default     = 5432
}

variable "db_name_app" {
  description = "Primary application database name"
  type        = string
}

variable "db_name_store" {
  description = "Whatsmeow store database name"
  type        = string
}

variable "redis_host" {
  description = "Redis endpoint hostname"
  type        = string
}

variable "redis_port" {
  description = "Redis port"
  type        = number
  default     = 6379
}

variable "redis_tls_enabled" {
  description = "Enable TLS connection to Redis"
  type        = bool
  default     = true
}

variable "s3_bucket_name" {
  description = "S3 bucket for media storage"
  type        = string
}

variable "s3_bucket_arn" {
  description = "S3 bucket ARN for IAM policies"
  type        = string
}

variable "s3_endpoint" {
  description = "Custom S3 endpoint (leave empty for AWS S3)"
  type        = string
  default     = ""
}

variable "s3_use_presigned_urls" {
  description = "Enable presigned URLs"
  type        = bool
  default     = true
}

variable "app_environment" {
  description = "Application environment value"
  type        = string
  default     = "production"
}

variable "log_level" {
  description = "Log level"
  type        = string
  default     = "info"
}

variable "extra_environment" {
  description = "Additional environment variables for API container"
  type        = map(string)
  default     = {}
}

variable "secret_key_mapping" {
  description = "Mapping of environment variable names to secret JSON keys"
  type        = map(string)
  default     = {}
}

variable "assign_public_ip" {
  description = "Assign public IP to ECS tasks"
  type        = bool
  default     = false
}

variable "api_image" {
  description = "Docker image for API container"
  type        = string
  default     = "whatsmeow-api:latest"
}

variable "task_cpu" {
  description = "Task CPU units (1024 = 1 vCPU)"
  type        = number
  default     = 1536 # 1.5 vCPU

  validation {
    condition     = contains([256, 512, 1024, 1536, 2048, 4096], var.task_cpu)
    error_message = "Task CPU must be a valid Fargate value."
  }
}

variable "task_memory" {
  description = "Task memory in MB"
  type        = number
  default     = 3072 # 3 GB

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
  description = "Enable ECS Exec for debugging (aws ecs execute-command)"
  type        = bool
  default     = false
}

variable "enable_autoscaling" {
  description = "Enable auto-scaling based on CPU/Memory"
  type        = bool
  default     = true
}

variable "autoscaling_min_capacity" {
  description = "Minimum number of tasks (when auto-scaling enabled)"
  type        = number
  default     = 1
}

variable "autoscaling_max_capacity" {
  description = "Maximum number of tasks (when auto-scaling enabled)"
  type        = number
  default     = 10
}

variable "autoscaling_cpu_target" {
  description = "Target CPU utilization percentage for auto-scaling"
  type        = number
  default     = 70

  validation {
    condition     = var.autoscaling_cpu_target >= 10 && var.autoscaling_cpu_target <= 90
    error_message = "CPU target must be between 10% and 90%."
  }
}

variable "autoscaling_memory_target" {
  description = "Target memory utilization percentage for auto-scaling"
  type        = number
  default     = 80

  validation {
    condition     = var.autoscaling_memory_target >= 10 && var.autoscaling_memory_target <= 90
    error_message = "Memory target must be between 10% and 90%."
  }
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
