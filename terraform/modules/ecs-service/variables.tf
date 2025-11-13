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

variable "redis_lock_key_prefix" {
  description = "Prefix applied to Redis distributed lock keys"
  type        = string
  default     = "funnelchat"
}

variable "redis_lock_ttl" {
  description = "TTL applied to Redis locks (duration string)"
  type        = string
  default     = "30s"
}

variable "redis_lock_refresh_interval" {
  description = "Interval between lock refresh attempts (duration string)"
  type        = string
  default     = "10s"
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

variable "api_base_url" {
  description = "Base URL for the API (used for OpenAPI documentation servers)"
  type        = string
  default     = ""
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

variable "prometheus_namespace" {
  description = "Prometheus metrics namespace exposed by the service"
  type        = string
  default     = "whatsmeow_api"
}

variable "media_local_storage_path" {
  description = "Filesystem path used for local media fallback storage"
  type        = string
  default     = "/tmp/whatsmeow/media"
}

variable "worker_heartbeat_interval" {
  description = "Interval used by each worker to send heartbeats (duration string)"
  type        = string
  default     = "5s"
}

variable "worker_heartbeat_expiry" {
  description = "Time window after which a worker heartbeat is considered expired"
  type        = string
  default     = "20s"
}

variable "worker_rebalance_interval" {
  description = "Interval for registry ownership rebalance checks"
  type        = string
  default     = "30s"
}

variable "sentry_release" {
  description = "Release identifier reported to Sentry"
  type        = string
  default     = "unknown"
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
