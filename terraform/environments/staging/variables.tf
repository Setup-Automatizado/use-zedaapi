# ==================================================
# Staging Environment - Variables
# ==================================================

# --------------------------------------------------
# Global
# --------------------------------------------------
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.1.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway"
  type        = bool
  default     = true
}

variable "certificate_arn" {
  description = "ACM certificate ARN for HTTPS (optional)"
  type        = string
  default     = null
}

# --------------------------------------------------
# Application Image & Runtime
# --------------------------------------------------
variable "api_image" {
  description = "Docker image URI for API service"
  type        = string
  default     = "whatsapp-api:staging"
}

variable "app_environment" {
  description = "Application environment string exposed via APP_ENV"
  type        = string
  default     = "staging"
}

variable "log_level" {
  description = "Structured log level"
  type        = string
  default     = "info"
}

variable "extra_environment" {
  description = "Additional environment variables for the API container"
  type        = map(string)
  default     = {}
}

variable "secret_env_mapping" {
  description = "Additional secret env var -> secret key mappings"
  type        = map(string)
  default     = {}
}

# --------------------------------------------------
# Database (Amazon RDS for PostgreSQL)
# --------------------------------------------------
variable "db_user" {
  description = "Database master username"
  type        = string
  default     = "whatsmeow"
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

variable "db_name_app" {
  description = "Primary application database name"
  type        = string
  default     = "api_core"
}

variable "db_name_store" {
  description = "Whatsmeow store database name"
  type        = string
  default     = "whatsmeow_store"
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.medium"
}

variable "db_allocated_storage" {
  description = "Initial allocated storage (GB)"
  type        = number
  default     = 20
}

variable "db_max_allocated_storage" {
  description = "Maximum storage for autoscaling (GB)"
  type        = number
  default     = 100
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
  description = "Backup retention period (days)"
  type        = number
  default     = 7
}

variable "db_deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

variable "db_skip_final_snapshot" {
  description = "Skip final snapshot on destroy"
  type        = bool
  default     = true
}

variable "db_apply_immediately" {
  description = "Apply RDS changes immediately"
  type        = bool
  default     = false
}

variable "db_performance_insights" {
  description = "Enable Performance Insights"
  type        = bool
  default     = true
}

variable "db_performance_insights_retention" {
  description = "Performance Insights retention (days)"
  type        = number
  default     = 7
}

# --------------------------------------------------
# Redis (Amazon ElastiCache)
# --------------------------------------------------
variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.1"
}

variable "redis_node_type" {
  description = "Cache node type"
  type        = string
  default     = "cache.t4g.small"
}

variable "redis_replicas_per_node_group" {
  description = "Replicas per shard"
  type        = number
  default     = 1
}

variable "redis_auth_token" {
  description = "Redis AUTH token (leave empty to disable AUTH)"
  type        = string
  default     = ""
  sensitive   = true
}

# --------------------------------------------------
# Object Storage (Amazon S3)
# --------------------------------------------------
variable "s3_bucket_name" {
  description = "S3 bucket name for media"
  type        = string
  default     = "whatsapp-api-staging-media"
}

variable "s3_force_destroy" {
  description = "Force destroy bucket even if not empty"
  type        = bool
  default     = false
}

variable "s3_endpoint" {
  description = "Custom S3 endpoint (optional)"
  type        = string
  default     = ""
}

variable "s3_use_presigned_urls" {
  description = "Enable presigned URLs for media access"
  type        = bool
  default     = true
}

variable "s3_lifecycle_rules" {
  description = "Lifecycle rules for media bucket"
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

# --------------------------------------------------
# Secrets
# --------------------------------------------------
variable "additional_secret_values" {
  description = "Additional key/value pairs stored in Secrets Manager"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "secret_recovery_window" {
  description = "Secrets Manager recovery window (days)"
  type        = number
  default     = 7
}

# --------------------------------------------------
# ECS Task Sizing & Scaling
# --------------------------------------------------
variable "task_cpu" {
  description = "Fargate task CPU units"
  type        = number
  default     = 1024
}

variable "task_memory" {
  description = "Fargate task memory (MB)"
  type        = number
  default     = 2048
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 1
}

variable "enable_execute_command" {
  description = "Enable ECS Exec"
  type        = bool
  default     = true
}

variable "enable_autoscaling" {
  description = "Enable ECS service autoscaling"
  type        = bool
  default     = true
}

variable "autoscaling_min_capacity" {
  description = "Minimum number of tasks"
  type        = number
  default     = 1
}

variable "autoscaling_max_capacity" {
  description = "Maximum number of tasks"
  type        = number
  default     = 5
}

variable "autoscaling_cpu_target" {
  description = "CPU utilization target for scaling"
  type        = number
  default     = 70
}

variable "autoscaling_memory_target" {
  description = "Memory utilization target for scaling"
  type        = number
  default     = 80
}

