# Staging environment configuration

aws_region = "us-east-1"

vpc_cidr           = "10.1.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = true

# Optional: uncomment when ACM certificate is issued in us-east-1
# certificate_arn = "arn:aws:acm:us-east-1:AKIA4W5HKAR23AI5S6VZ:certificate/CERTIFICATE_ID"

api_image       = "873839854709.dkr.ecr.us-east-1.amazonaws.com/whatsapp-api:staging"
app_environment = "staging"
log_level       = "info"

db_user     = "whatsmeow"
db_password = "joXuRHCec93TsM1X"

db_instance_class        = "db.t4g.medium"
db_allocated_storage     = 20
db_max_allocated_storage = 100
db_multi_az              = false

redis_engine_version          = "7.1"
redis_node_type               = "cache.t4g.small"
redis_replicas_per_node_group = 1
redis_auth_token              = "joXuRHCec93TsM1X"
redis_lock_key_prefix         = "funnelchat"
redis_lock_ttl                = "30s"
redis_lock_refresh_interval   = "10s"

s3_bucket_name        = "staging-whatsapp-api-media"
s3_force_destroy      = false
s3_use_presigned_urls = true
s3_access_key         = "AKIA4W5HKAR2XE36XHBI"
s3_secret_key         = "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"
s3_public_base_url    = "https://staging-whatsapp-api-media.s3.us-east-1.amazonaws.com"
api_base_url          = "http://staging-whatsmeow-alb-1412624585.us-east-1.elb.amazonaws.com"

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
  partner_auth_token = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
  client_auth_token  = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
  sentry_dsn         = "https://CHANGE_ME_STAGING_SENTRY_DSN"
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

  # Proxy Pool Configuration (Webshare Integration)
  PROXY_POOL_ENABLED                = "true"
  PROXY_POOL_SYNC_INTERVAL          = "5m"
  PROXY_POOL_DEFAULT_COUNTRY_CODES  = "BR,AR,CL,CO,MX,PE,EC,UY,PY,BO,VE,US,CA"
  PROXY_POOL_ASSIGNMENT_RETRY_DELAY = "5s"
  PROXY_POOL_MAX_ASSIGNMENT_RETRIES = "3"
  PROXY_POOL_WEBSHARE_ENDPOINT      = "https://proxy.webshare.io/api/v2"
  PROXY_POOL_WEBSHARE_PLAN_ID       = "12784271"
  PROXY_POOL_WEBSHARE_MODE          = "direct"
}

media_local_secret_key      = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
media_local_public_base_url = "http://staging-whatsmeow-alb-1412624585.us-east-1.elb.amazonaws.com"
redis_username              = "default"
worker_heartbeat_interval   = "5s"
worker_heartbeat_expiry     = "20s"
worker_rebalance_interval   = "30s"
