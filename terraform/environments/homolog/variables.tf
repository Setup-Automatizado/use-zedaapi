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

variable "redis_lock_key_prefix" {
  description = "Prefix applied to Redis lock keys"
  type        = string
  default     = "funnelchat"
}

variable "redis_lock_ttl" {
  description = "TTL applied to Redis locks"
  type        = string
  default     = "30s"
}

variable "redis_lock_refresh_interval" {
  description = "Interval between lock refresh attempts"
  type        = string
  default     = "10s"
}

variable "media_local_storage_path" {
  description = "Path used by ECS task for local media fallback"
  type        = string
  default     = "/tmp/whatsmeow/media"
}

variable "worker_heartbeat_interval" {
  description = "Interval between worker heartbeats"
  type        = string
  default     = "5s"
}

variable "worker_heartbeat_expiry" {
  description = "Time window before a worker is considered unhealthy"
  type        = string
  default     = "20s"
}

variable "worker_rebalance_interval" {
  description = "Interval for the ownership rebalance loop"
  type        = string
  default     = "30s"
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

variable "rds_publicly_accessible" {
  description = "Expose the RDS instance with a public endpoint (only for non-production environments)"
  type        = bool
  default     = false
}

variable "rds_use_public_subnets" {
  description = "Place the RDS subnet group in the public subnets instead of private ones"
  type        = bool
  default     = false
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

variable "allowed_admin_ips" {
  description = "List of IP addresses allowed to access RDS directly (for admin/database management tasks)"
  type        = list(string)
  default     = []
}

# ==================================================
# Manager Frontend Variables
# ==================================================

variable "enable_manager" {
  description = "Enable Manager frontend service"
  type        = bool
  default     = false
}

variable "manager_image" {
  description = "Docker image URI for Manager"
  type        = string
  default     = ""
}

variable "manager_app_url" {
  description = "Public URL for Manager application"
  type        = string
  default     = ""
}

variable "manager_host_header" {
  description = "Host header for Manager routing (empty for path-based routing)"
  type        = string
  default     = ""
}

variable "manager_task_cpu" {
  description = "Manager task CPU"
  type        = number
  default     = 256
}

variable "manager_task_memory" {
  description = "Manager task memory"
  type        = number
  default     = 512
}

variable "manager_desired_count" {
  description = "Manager desired task count"
  type        = number
  default     = 1
}

variable "manager_db_name" {
  description = "Manager database name"
  type        = string
  default     = "manager_db"
}

variable "manager_smtp_host" {
  description = "SMTP host for Manager emails"
  type        = string
  default     = ""
}

variable "manager_smtp_port" {
  description = "SMTP port for Manager emails"
  type        = number
  default     = 587
}

variable "manager_smtp_user" {
  description = "SMTP username for Manager emails"
  type        = string
  default     = ""
}

variable "manager_email_from_name" {
  description = "Email sender display name"
  type        = string
  default     = "WhatsApp Manager"
}

variable "manager_email_from_address" {
  description = "Email sender address"
  type        = string
  default     = "noreply@example.com"
}

variable "manager_support_email" {
  description = "Support email address"
  type        = string
  default     = "support@example.com"
}

variable "manager_additional_secrets" {
  description = "Additional secrets for Manager (better_auth_secret, smtp_password)"
  type        = map(string)
  default     = {}
  sensitive   = true
}

# --------------------------------------------------
# Manager OAuth Configuration
# --------------------------------------------------

variable "manager_github_client_id" {
  description = "GitHub OAuth Client ID for Manager"
  type        = string
  default     = ""
}

variable "manager_google_client_id" {
  description = "Google OAuth Client ID for Manager"
  type        = string
  default     = ""
}

# --------------------------------------------------
# Manager S3/MinIO Override Configuration
# --------------------------------------------------

variable "manager_s3_endpoint" {
  description = "Custom S3/MinIO endpoint for Manager (overrides default)"
  type        = string
  default     = ""
}

variable "manager_s3_bucket" {
  description = "Custom S3 bucket for Manager"
  type        = string
  default     = ""
}

variable "manager_s3_public_url" {
  description = "Custom S3 public URL for Manager"
  type        = string
  default     = ""
}
