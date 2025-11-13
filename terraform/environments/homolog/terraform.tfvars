# Homolog environment configuration

aws_region = "us-east-1"

vpc_cidr           = "10.2.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = false

# certificate_arn = "arn:aws:acm:us-east-1:AKIA4W5HKAR23AI5S6VZ:certificate/CERTIFICATE_ID"

api_image       = "873839854709.dkr.ecr.us-east-1.amazonaws.com/whatsapp-api:homolog"
app_environment = "homolog"
log_level       = "debug"

db_user     = "whatsmeow"
db_password = "joXuRHCec93TsM1X"

db_instance_class        = "db.t4g.small"
db_allocated_storage     = 10
db_max_allocated_storage = 50
db_multi_az              = false
db_engine_version        = "16.8"
db_backup_retention      = 3

redis_engine_version          = "7.1"
redis_node_type               = "cache.t4g.small"
redis_replicas_per_node_group = 0
redis_auth_token              = ""
redis_lock_key_prefix         = "funnelchat"
redis_lock_ttl                = "30s"
redis_lock_refresh_interval   = "10s"

s3_bucket_name        = "homolog-whatsapp-api-media"
s3_force_destroy      = true
s3_use_presigned_urls = true
s3_access_key         = "AKIA4W5HKAR2XE36XHBI"
s3_secret_key         = "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"
s3_public_base_url    = "https://homolog-whatsapp-api-media.s3.us-east-1.amazonaws.com"
api_base_url          = "http://homolog-whatsmeow-alb-1665982637.us-east-1.elb.amazonaws.com"

media_local_secret_key      = "80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
media_local_public_base_url = "http://homolog-whatsmeow-alb-1665982637.us-east-1.elb.amazonaws.com/media"
redis_username              = "default"
worker_heartbeat_interval   = "5s"
worker_heartbeat_expiry     = "20s"
worker_rebalance_interval   = "30s"

s3_lifecycle_rules = []

additional_secret_values = {
  partner_auth_token = "joXuRHCec93TsM1X",
  client_auth_token  = "joXuRHCec93TsM1X",
}

secret_env_mapping = {
  PARTNER_AUTH_TOKEN     = "partner_auth_token"
  CLIENT_AUTH_TOKEN      = "client_auth_token"
  S3_ACCESS_KEY          = "s3_access_key"
  S3_SECRET_KEY          = "s3_secret_key"
  MEDIA_LOCAL_SECRET_KEY = "media_local_secret_key"
}

secret_recovery_window = 7

task_cpu               = 512
task_memory            = 1024
desired_count          = 1
enable_execute_command = true

enable_autoscaling        = false
autoscaling_min_capacity  = 1
autoscaling_max_capacity  = 2
autoscaling_cpu_target    = 75
autoscaling_memory_target = 85

extra_environment = {
  LOG_LEVEL = "debug"
}

# Admin IPs allowed to access RDS directly (for database management)
# ⚠️ WARNING: Using 0.0.0.0/0 allows access from ANY IP - INSECURE!
# Only use in homolog/development environments temporarily
allowed_admin_ips = ["0.0.0.0/0"]
