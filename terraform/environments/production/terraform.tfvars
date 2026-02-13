# Production environment configuration

aws_region = "us-east-1"

vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
enable_nat_gateway = true

# CloudFront handles HTTPS - no ACM certificate needed for ALB
certificate_arn = null

api_image       = "873839854709.dkr.ecr.us-east-1.amazonaws.com/whatsapp-api:latest"
app_environment = "production"
log_level       = "info"

db_user     = "whatsmeow"
db_password = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"

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
redis_auth_token              = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
redis_lock_key_prefix         = "funnelchat"
redis_lock_ttl                = "30s"
redis_lock_refresh_interval   = "10s"

s3_bucket_name        = "production-whatsapp-api-media"
s3_force_destroy      = false
s3_use_presigned_urls = true
s3_access_key         = "AKIA4W5HKAR2XE36XHBI"
s3_secret_key         = "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"
s3_public_base_url    = "https://production-whatsapp-api-media.s3.us-east-1.amazonaws.com"
api_base_url          = "https://d2qqtmk57jbmfv.cloudfront.net"

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
  partner_auth_token = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
  client_auth_token  = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
  sentry_dsn         = "https://2e9b95288cbd7051e60d263580ceb8e0@o4510360009441280.ingest.us.sentry.io/4510360011538432"
  webshare_api_key   = "shhz7zca1xjtkazd2do791kbyrmfy0wkqbtkxzuy"
}

secret_env_mapping = {
  PARTNER_AUTH_TOKEN          = "partner_auth_token"
  CLIENT_AUTH_TOKEN           = "client_auth_token"
  SENTRY_DSN                 = "sentry_dsn"
  S3_ACCESS_KEY              = "s3_access_key"
  S3_SECRET_KEY              = "s3_secret_key"
  MEDIA_LOCAL_SECRET_KEY     = "media_local_secret_key"
  PROXY_POOL_WEBSHARE_API_KEY = "webshare_api_key"
}

secret_recovery_window = 14

task_cpu               = 2048
task_memory            = 4096
desired_count          = 2
enable_execute_command = false

enable_autoscaling        = true
autoscaling_min_capacity  = 2
autoscaling_max_capacity  = 6
autoscaling_cpu_target    = 60
autoscaling_memory_target = 70

extra_environment = {
  PROMETHEUS_ENABLED                = "true"
  GIN_MODE                         = "release"
  API_ECHO_ENABLED                 = "true"
  FILTER_WAITING_MESSAGE           = "true"
  FILTER_SECONDARY_DEVICE_RECEIPTS = "true"

  # Status Cache (Redis-based webhook deduplication)
  STATUS_CACHE_ENABLED = "true"

  # Proxy Pool Configuration (Webshare Integration)
  PROXY_POOL_ENABLED                = "true"
  PROXY_POOL_SYNC_INTERVAL          = "5m"
  PROXY_POOL_DEFAULT_COUNTRY_CODES  = "BR,US,AR,CL,CO,MX,PE,EC,UY,PY,BO,VE,US,CA"
  PROXY_POOL_ASSIGNMENT_RETRY_DELAY = "5s"
  PROXY_POOL_MAX_ASSIGNMENT_RETRIES = "3"
  PROXY_POOL_WEBSHARE_ENDPOINT      = "https://proxy.webshare.io/api/v2"
  PROXY_POOL_WEBSHARE_PLAN_ID       = "12784271"
  PROXY_POOL_WEBSHARE_MODE          = "direct"
}

media_local_secret_key      = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
media_local_public_base_url = "https://d2qqtmk57jbmfv.cloudfront.net"
redis_username              = "default"
worker_heartbeat_interval   = "5s"
worker_heartbeat_expiry     = "20s"
worker_rebalance_interval   = "30s"

# --------------------------------------------------
# Manager Frontend Configuration
# --------------------------------------------------
enable_manager              = true
manager_image               = "873839854709.dkr.ecr.us-east-1.amazonaws.com/manager-whatsapp-api:latest"
manager_app_url             = "https://d1zkc6ehods2po.cloudfront.net"
manager_task_cpu            = 512
manager_task_memory         = 1024
manager_desired_count       = 2
manager_db_name             = "manager_db"

# SMTP Configuration for Manager
manager_smtp_host           = "smtp-relay.brevo.com"
manager_smtp_port           = 587
manager_smtp_user           = "6a53de001@smtp-brevo.com"
manager_email_from_name     = "WhatsApp Manager"
manager_email_from_address  = "noreply@setupautomatizado.com.br"
manager_support_email       = "suporte@setupautomatizado.com.br"

# S3 Configuration for Manager
manager_s3_endpoint   = ""
manager_s3_bucket     = ""
manager_s3_public_url = ""

# OAuth Configuration for Manager
manager_github_client_id = ""
manager_google_client_id = ""

# Manager Secrets (override via environment or tfvars.local)
manager_additional_secrets = {
  better_auth_secret = "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
  smtp_password      = "xsmtpsib-6f1c61ab72982976c2b0f00be2876a16f59dda8be94bd25070aeddd4dd205eba-1YJO7T03Q6dxFyVZ"
}
