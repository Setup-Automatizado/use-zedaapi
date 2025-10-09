# Production environment configuration

aws_region = "us-east-1"

vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
enable_nat_gateway = true

certificate_arn = "arn:aws:acm:us-east-1:ACCOUNT_ID:certificate/CERTIFICATE_ID"

api_image       = "ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/whatsapp-api:prod"
app_environment = "production"
log_level       = "info"

db_user     = "whatsmeow"
db_password = "CHANGE_ME_PRODUCTION_DB_PASSWORD"

db_instance_class        = "db.r6g.large"
db_allocated_storage     = 100
db_max_allocated_storage = 500
db_multi_az              = true
db_backup_retention      = 14
db_deletion_protection   = true
db_skip_final_snapshot   = false

redis_engine_version          = "7.1"
redis_node_type               = "cache.r6g.large"
redis_replicas_per_node_group = 2
redis_auth_token              = "CHANGE_ME_PRODUCTION_REDIS_TOKEN"

s3_bucket_name        = "production-whatsapp-api-media"
s3_force_destroy      = false
s3_use_presigned_urls = true

s3_lifecycle_rules = [
  {
    id      = "archive-180d"
    enabled = true
    transitions = [
      {
        days          = 180
        storage_class = "STANDARD_IA"
      },
      {
        days          = 365
        storage_class = "GLACIER"
      }
    ]
    expiration_days = 1095
  }
]

additional_secret_values = {
  partner_auth_token = "CHANGE_ME_PRODUCTION_PARTNER_TOKEN"
  sentry_dsn         = "https://CHANGE_ME_PRODUCTION_SENTRY_DSN"
}

secret_env_mapping = {
  PARTNER_AUTH_TOKEN = "partner_auth_token"
  SENTRY_DSN         = "sentry_dsn"
}

secret_recovery_window = 14

task_cpu               = 2048
task_memory            = 4096
desired_count          = 2
enable_execute_command = false

enable_autoscaling         = true
autoscaling_min_capacity   = 2
autoscaling_max_capacity   = 6
autoscaling_cpu_target     = 60
autoscaling_memory_target  = 70

extra_environment = {
  PROMETHEUS_ENABLED = "true"
  GIN_MODE           = "release"
}
