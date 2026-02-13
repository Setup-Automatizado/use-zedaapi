# ==================================================
# WhatsApp API - Homolog Environment Configuration
# ==================================================
# Este arquivo contém apenas variáveis suportadas pelo
# módulo raiz (ver variables.tf). Ajuste os valores
# conforme necessidade do ambiente antes de executar
# `terraform plan`/`terraform apply`.
# ==================================================

# --------------------------------------------------
# Região, rede e segurança
# --------------------------------------------------
aws_region         = "us-east-1"
availability_zones = ["us-east-1a", "us-east-1b"]
vpc_cidr           = "10.2.0.0/16"
enable_nat_gateway = false

# IPs com acesso administrativo direto a RDS/Redis
allowed_admin_ips = [
  "186.205.11.185/32",
]

# --------------------------------------------------
# Aplicação e observabilidade
# --------------------------------------------------
app_environment      = "homolog"
log_level            = "debug"
prometheus_namespace = "whatsmeow_api"
sentry_release       = "homolog"

# --------------------------------------------------
# Banco de dados (RDS PostgreSQL)
# --------------------------------------------------
db_name_app              = "whatsmeow"
db_name_store            = "whatsmeow_store"
db_user                  = "whatsapp_api_user"
db_password              = "yjMZ(j(J(DPZUXZusvz6"
db_instance_class        = "db.t3.micro"
db_allocated_storage     = 20
db_max_allocated_storage = 50
db_engine_version        = "16.8"
db_multi_az              = false
db_backup_retention      = 1
db_deletion_protection   = false
db_skip_final_snapshot   = true
db_apply_immediately     = true
db_performance_insights  = false

rds_publicly_accessible = true
rds_use_public_subnets  = true

# --------------------------------------------------
# Redis (ElastiCache)
# --------------------------------------------------
redis_engine_version          = "7.1"
redis_node_type               = "cache.t3.micro"
redis_replicas_per_node_group = 0
redis_auth_token              = "y52kar3uAQuu5TDgt27stBfvPcoaSwh1"
redis_lock_key_prefix         = "funnelchat"
redis_lock_ttl                = "30s"
redis_lock_refresh_interval   = "10s"

# --------------------------------------------------
# S3 (armazenamento de mídia)
# --------------------------------------------------
s3_bucket_name   = "whatsapp-api-homolog-media"
s3_force_destroy = true
s3_lifecycle_rules = [
  {
    id      = "media-retention"
    enabled = true
    transitions = [
      {
        days          = 30
        storage_class = "STANDARD_IA"
      }
    ]
    expiration_days = 90
  }
]

s3_public_base_url          = "https://whatsapp-api-homolog-media.s3.us-east-1.amazonaws.com"
media_local_public_base_url = "http://homolog-whatsmeow-alb-731186848.us-east-1.elb.amazonaws.com/media"
s3_use_presigned_urls       = true

# --------------------------------------------------
# Secrets Manager payload extra
# --------------------------------------------------
# S3 credentials vazias para usar IAM Role do ECS Task
# Em dev local com MinIO, defina credenciais no .env
s3_access_key          = ""
s3_secret_key          = ""
media_local_secret_key = "supersecret"
additional_secret_values = {
  partner_auth_token = "iB4uIxOYOMFSnScXWlphBg=="
  client_auth_token  = "iB4uIxOYOMFSnScXWlphBg=="
  sentry_dsn         = "https://2e9b95288cbd7051e60d263580ceb8e0@o4510360009441280.ingest.us.sentry.io/4510360011538432"
  webshare_api_key   = "shhz7zca1xjtkazd2do791kbyrmfy0wkqbtkxzuy"
}

# --------------------------------------------------
# ECS / serviço API
# --------------------------------------------------
api_image                 = "873839854709.dkr.ecr.us-east-1.amazonaws.com/whatsapp-api:homolog"
task_cpu                  = 512
task_memory               = 1024
desired_count             = 1
enable_execute_command    = true
enable_autoscaling        = true
autoscaling_min_capacity  = 1
autoscaling_max_capacity  = 2
autoscaling_cpu_target    = 70
autoscaling_memory_target = 85

# --------------------------------------------------
# Certificado do ALB (HTTP somente em homolog)
# --------------------------------------------------
certificate_arn = null

# --------------------------------------------------
# Outros ajustes
# --------------------------------------------------
media_local_storage_path  = "/tmp/whatsmeow/media"
worker_heartbeat_interval = "5s"
worker_heartbeat_expiry   = "20s"
worker_rebalance_interval = "30s"
secret_recovery_window    = 7
s3_endpoint               = ""
api_base_url              = "http://homolog-whatsmeow-alb-731186848.us-east-1.elb.amazonaws.com"
extra_environment = {
  # Application Feature Flags
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
secret_env_mapping = {
  PARTNER_AUTH_TOKEN          = "partner_auth_token"
  CLIENT_AUTH_TOKEN           = "client_auth_token"
  SENTRY_DSN                 = "sentry_dsn"
  S3_ACCESS_KEY              = "s3_access_key"
  S3_SECRET_KEY              = "s3_secret_key"
  MEDIA_LOCAL_SECRET_KEY     = "media_local_secret_key"
  PROXY_POOL_WEBSHARE_API_KEY = "webshare_api_key"
}
redis_username = "default"

# --------------------------------------------------
# Manager Frontend Configuration
# --------------------------------------------------
enable_manager              = true
manager_image               = "873839854709.dkr.ecr.us-east-1.amazonaws.com/manager-whatsapp-api:homolog"
manager_app_url             = "http://homolog-whatsmeow-alb-731186848.us-east-1.elb.amazonaws.com"
manager_host_header         = ""  # Path-based routing (Manager as default)
manager_task_cpu            = 256
manager_task_memory         = 512
manager_desired_count       = 1
manager_db_name             = "manager_db"

# SMTP Configuration for Manager
manager_smtp_host           = "smtp-relay.brevo.com"
manager_smtp_port           = 587
manager_smtp_user           = "6a53de001@smtp-brevo.com"
manager_email_from_name     = "WhatsApp Manager"
manager_email_from_address  = "noreply@setupautomatizado.com.br"
manager_support_email       = "suporte@setupautomatizado.com.br"

# S3/MinIO Configuration for Manager (external MinIO)
manager_s3_endpoint   = "https://s3.setupautomatizado.com.br"
manager_s3_bucket     = "funnelchat"
manager_s3_public_url = "https://s3.setupautomatizado.com.br/insightzap"

# OAuth Configuration for Manager
manager_github_client_id = "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"
manager_google_client_id = "57219513982-t59t8r298t0tngihig5fvc54i0huh0c0.apps.googleusercontent.com"

# Manager Secrets (sensitive - override via environment or tfvars.local)
manager_additional_secrets = {
  better_auth_secret = "s2fVn34EJ3dEWt9cHHrltPHipfpCBsqE"
  smtp_password      = "xsmtpsib-6f1c61ab72982976c2b0f00be2876a16f59dda8be94bd25070aeddd4dd205eba-1YJO7T03Q6dxFyVZ"
}
