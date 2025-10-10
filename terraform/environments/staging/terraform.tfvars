# Staging environment configuration

aws_region = "us-east-1"

vpc_cidr           = "10.1.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = true

# Optional: uncomment when ACM certificate is issued in us-east-1
# certificate_arn = "arn:aws:acm:us-east-1:ACCOUNT_ID:certificate/CERTIFICATE_ID"

api_image       = "ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/whatsapp-api:staging"
app_environment = "staging"
log_level       = "info"

db_user     = "whatsmeow"
db_password = "CHANGE_ME_STAGING_DB_PASSWORD"

db_instance_class        = "db.t4g.medium"
db_allocated_storage     = 20
db_max_allocated_storage = 100
db_multi_az              = false

redis_engine_version          = "7.1"
redis_node_type               = "cache.t4g.small"
redis_replicas_per_node_group = 1
redis_auth_token              = "CHANGE_ME_STAGING_REDIS_TOKEN"

s3_bucket_name        = "staging-whatsapp-api-media"
s3_force_destroy      = false
s3_use_presigned_urls = true
s3_access_key         = ""
s3_secret_key         = ""
s3_public_base_url    = ""

# Example lifecycle rule; adjust or remove as needed
s3_lifecycle_rules = [
  {
    id      = "archive-old-media"
    enabled = true
    transitions = [
      {
        days          = 90
        storage_class = "STANDARD_IA"
      }
    ]
    expiration_days = 365
  }
]

additional_secret_values = {
  partner_auth_token = "CHANGE_ME_STAGING_PARTNER_TOKEN"
  sentry_dsn         = "https://CHANGE_ME_STAGING_SENTRY_DSN"
}

secret_env_mapping = {
  PARTNER_AUTH_TOKEN     = "partner_auth_token"
  SENTRY_DSN             = "sentry_dsn"
  S3_ACCESS_KEY          = "s3_access_key"
  S3_SECRET_KEY          = "s3_secret_key"
  MEDIA_LOCAL_SECRET_KEY = "media_local_secret_key"
}

secret_recovery_window = 7

task_cpu               = 1024
task_memory            = 2048
desired_count          = 1
enable_execute_command = true

enable_autoscaling        = true
autoscaling_min_capacity  = 1
autoscaling_max_capacity  = 5
autoscaling_cpu_target    = 70
autoscaling_memory_target = 80

extra_environment = {
  PROMETHEUS_ENABLED = "true"
}

media_local_secret_key      = ""
media_local_public_base_url = ""
redis_username              = ""
