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
  "177.192.10.4/32",
  "102.216.82.164/32",
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
extra_environment         = {}
secret_env_mapping        = {}
