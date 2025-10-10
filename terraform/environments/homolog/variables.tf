# ==================================================
# Homolog Environment - Variables
# ==================================================

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.2.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway"
  type        = bool
  default     = false
}

variable "certificate_arn" {
  description = "ACM certificate ARN (optional)"
  type        = string
  default     = null
}

variable "api_image" {
  description = "Docker image URI"
  type        = string
  default     = "whatsapp-api:homolog"
}

variable "app_environment" {
  description = "Application environment string"
  type        = string
  default     = "homolog"
}

variable "log_level" {
  description = "Structured log level"
  type        = string
  default     = "debug"
}

variable "prometheus_namespace" {
  description = "Prometheus metrics namespace"
  type        = string
  default     = "whatsmeow_api"
}

variable "sentry_release" {
  description = "Sentry release identifier"
  type        = string
  default     = "homolog"
}

variable "extra_environment" {
  description = "Additional environment variables"
  type        = map(string)
  default     = {}
}

variable "secret_env_mapping" {
  description = "Additional secret env mappings"
  type        = map(string)
  default     = {}
}

variable "s3_access_key" {
  description = "S3 access key for presigned URLs"
  type        = string
  default     = ""
  sensitive   = true
}

variable "s3_secret_key" {
  description = "S3 secret key for presigned URLs"
  type        = string
  default     = ""
  sensitive   = true
}

variable "s3_public_base_url" {
  description = "Public base URL for S3-accessible media"
  type        = string
  default     = ""
}

variable "media_local_secret_key" {
  description = "Secret key used to sign local media URLs"
  type        = string
  default     = ""
  sensitive   = true
}

variable "media_local_public_base_url" {
  description = "Public base URL for locally-served media"
  type        = string
  default     = ""
}

variable "api_base_url" {
  description = "Base URL for the API (used for OpenAPI documentation servers)"
  type        = string
  default     = ""
}

variable "redis_username" {
  description = "Redis username (if ACLs are enabled)"
  type        = string
  default     = ""
}

variable "media_local_storage_path" {
  description = "Path used by ECS task for local media fallback"
  type        = string
  default     = "/tmp/whatsmeow/media"
}

variable "db_user" {
  description = "Database user"
  type        = string
  default     = "whatsmeow"
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name_app" {
  description = "Primary database name"
  type        = string
  default     = "api_core"
}

variable "db_name_store" {
  description = "Store database name"
  type        = string
  default     = "whatsmeow_store"
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.small"
}

variable "db_allocated_storage" {
  description = "Allocated storage"
  type        = number
  default     = 10
}

variable "db_max_allocated_storage" {
  description = "Max storage"
  type        = number
  default     = 50
}

variable "db_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "16.3"
}

variable "db_multi_az" {
  description = "Enable Multi-AZ"
  type        = bool
  default     = false
}

variable "db_backup_retention" {
  description = "Backup retention"
  type        = number
  default     = 3
}

variable "db_deletion_protection" {
  description = "Deletion protection"
  type        = bool
  default     = false
}

variable "db_skip_final_snapshot" {
  description = "Skip final snapshot"
  type        = bool
  default     = true
}

variable "db_apply_immediately" {
  description = "Apply changes immediately"
  type        = bool
  default     = false
}

variable "db_performance_insights" {
  description = "Enable Performance Insights"
  type        = bool
  default     = false
}

variable "db_performance_insights_retention" {
  description = "Performance Insights retention"
  type        = number
  default     = 7
}

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.1"
}

variable "redis_node_type" {
  description = "Redis node type"
  type        = string
  default     = "cache.t4g.small"
}

variable "redis_replicas_per_node_group" {
  description = "Replicas per shard"
  type        = number
  default     = 0
}

variable "redis_auth_token" {
  description = "Redis AUTH token"
  type        = string
  default     = ""
  sensitive   = true
}

variable "s3_bucket_name" {
  description = "S3 bucket name"
  type        = string
  default     = "whatsapp-api-homolog-media"
}

variable "s3_force_destroy" {
  description = "Force destroy bucket"
  type        = bool
  default     = true
}

variable "s3_endpoint" {
  description = "Custom S3 endpoint"
  type        = string
  default     = ""
}

variable "s3_use_presigned_urls" {
  description = "Enable presigned URLs"
  type        = bool
  default     = true
}

variable "s3_lifecycle_rules" {
  description = "Lifecycle rules"
  type = list(object({
    id      = string
    enabled = bool
    transitions = optional(list(object({
      days          = number
      storage_class = string
    })), [])
    expiration_days = optional(number)
  }))
  default = []
}

variable "additional_secret_values" {
  description = "Additional secret key/value pairs"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "secret_recovery_window" {
  description = "Secrets recovery window"
  type        = number
  default     = 7
}

variable "task_cpu" {
  description = "Task CPU"
  type        = number
  default     = 512
}

variable "task_memory" {
  description = "Task memory"
  type        = number
  default     = 1024
}

variable "desired_count" {
  description = "Desired tasks"
  type        = number
  default     = 1
}

variable "enable_execute_command" {
  description = "Enable ECS Exec"
  type        = bool
  default     = true
}

variable "enable_autoscaling" {
  description = "Enable autoscaling"
  type        = bool
  default     = false
}

variable "autoscaling_min_capacity" {
  description = "Min tasks"
  type        = number
  default     = 1
}

variable "autoscaling_max_capacity" {
  description = "Max tasks"
  type        = number
  default     = 2
}

variable "autoscaling_cpu_target" {
  description = "CPU target"
  type        = number
  default     = 75
}

variable "autoscaling_memory_target" {
  description = "Memory target"
  type        = number
  default     = 85
}
