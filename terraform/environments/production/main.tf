# ==================================================
# Production Environment - Main Configuration
# ==================================================

terraform {
  required_version = ">= 1.9.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "whatsmeow-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "whatsmeow-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "WhatsApp API"
      Environment = "production"
      ManagedBy   = "Terraform"
      Repository  = "whatsapp-api-golang"
    }
  }
}

locals {
  environment    = "production"
  db_name_app    = var.db_name_app
  db_name_store  = var.db_name_store
  s3_bucket_name = var.s3_bucket_name

  common_tags = {
    Environment = local.environment
    Project     = "WhatsApp API"
  }
}

# --------------------------------------------------
# Networking
# --------------------------------------------------
module "vpc" {
  source = "../../modules/vpc"

  environment        = local.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  enable_nat_gateway = var.enable_nat_gateway
  enable_flow_logs   = true

  tags = local.common_tags
}

module "security_groups" {
  source = "../../modules/security-groups"

  environment = local.environment
  vpc_id      = module.vpc.vpc_id

  tags = local.common_tags
}

# --------------------------------------------------
# Load Balancing & Compute
# --------------------------------------------------
module "alb" {
  source = "../../modules/alb"

  environment                = local.environment
  vpc_id                     = module.vpc.vpc_id
  public_subnet_ids          = module.vpc.public_subnet_ids
  alb_security_group_id      = module.security_groups.alb_security_group_id
  certificate_arn            = var.certificate_arn
  expose_minio_console       = false
  enable_deletion_protection = true

  tags = local.common_tags
}

module "ecs_cluster" {
  source = "../../modules/ecs-cluster"

  environment               = local.environment
  enable_container_insights = true
  enable_fargate_spot       = false

  tags = local.common_tags
}

# --------------------------------------------------
# Data Stores
# --------------------------------------------------
module "s3" {
  source = "../../modules/s3"

  environment   = local.environment
  bucket_name   = local.s3_bucket_name
  force_destroy = var.s3_force_destroy

  lifecycle_rules = var.s3_lifecycle_rules
  tags            = local.common_tags
}

module "rds" {
  source = "../../modules/rds"

  environment                    = local.environment
  db_name                        = local.db_name_app
  db_username                    = var.db_user
  db_password                    = var.db_password
  allocated_storage              = var.db_allocated_storage
  max_allocated_storage          = var.db_max_allocated_storage
  instance_class                 = var.db_instance_class
  engine_version                 = var.db_engine_version
  multi_az                       = var.db_multi_az
  backup_retention_period        = var.db_backup_retention
  deletion_protection            = var.db_deletion_protection
  skip_final_snapshot            = var.db_skip_final_snapshot
  apply_immediately              = var.db_apply_immediately
  performance_insights_enabled   = var.db_performance_insights
  performance_insights_retention = var.db_performance_insights_retention
  subnet_ids                     = module.vpc.private_subnet_ids
  security_group_ids             = [module.security_groups.rds_security_group_id]

  tags = local.common_tags
}

module "elasticache" {
  source = "../../modules/elasticache"

  environment                = local.environment
  engine_version             = var.redis_engine_version
  node_type                  = var.redis_node_type
  replicas_per_node_group    = var.redis_replicas_per_node_group
  automatic_failover_enabled = true
  multi_az_enabled           = true
  auth_token                 = var.redis_auth_token
  subnet_ids                 = module.vpc.private_subnet_ids
  security_group_ids         = [module.security_groups.redis_security_group_id]

  tags = local.common_tags
}

# --------------------------------------------------
# Secrets
# --------------------------------------------------
locals {
  db_user_url         = urlencode(var.db_user)
  db_password_url     = urlencode(var.db_password)
  postgres_dsn        = "postgres://${local.db_user_url}:${local.db_password_url}@${module.rds.db_endpoint}:${module.rds.db_port}/${local.db_name_app}?sslmode=require"
  wameow_postgres_dsn = "postgres://${local.db_user_url}:${local.db_password_url}@${module.rds.db_endpoint}:${module.rds.db_port}/${local.db_name_store}?sslmode=require"

  secret_payload = merge({
    db_user                = var.db_user,
    db_password            = var.db_password,
    postgres_dsn           = local.postgres_dsn,
    wameow_postgres_dsn    = local.wameow_postgres_dsn,
    redis_password         = var.redis_auth_token,
    s3_access_key          = var.s3_access_key,
    s3_secret_key          = var.s3_secret_key,
    media_local_secret_key = var.media_local_secret_key
  }, var.additional_secret_values)
}

module "secrets" {
  source = "../../modules/secrets"

  environment             = local.environment
  secret_payload          = local.secret_payload
  recovery_window_in_days = var.secret_recovery_window

  tags = local.common_tags
}

# --------------------------------------------------
# ECS Service
# --------------------------------------------------
module "ecs_service" {
  source = "../../modules/ecs-service"

  environment               = local.environment
  cluster_id                = module.ecs_cluster.cluster_id
  cluster_name              = module.ecs_cluster.cluster_name
  private_subnet_ids        = module.vpc.private_subnet_ids
  ecs_security_group_id     = module.security_groups.ecs_tasks_security_group_id
  api_target_group_arn      = module.alb.api_target_group_arn
  secrets_arn               = module.secrets.secret_arn
  api_image                 = var.api_image
  task_cpu                  = var.task_cpu
  task_memory               = var.task_memory
  desired_count             = var.desired_count
  enable_execute_command    = var.enable_execute_command
  enable_autoscaling        = var.enable_autoscaling
  autoscaling_min_capacity  = var.autoscaling_min_capacity
  autoscaling_max_capacity  = var.autoscaling_max_capacity
  autoscaling_cpu_target    = var.autoscaling_cpu_target
  autoscaling_memory_target = var.autoscaling_memory_target
  aws_region                = var.aws_region
  db_host                   = module.rds.db_endpoint
  db_port                   = module.rds.db_port
  db_name_app               = local.db_name_app
  db_name_store             = local.db_name_store
  redis_host                = module.elasticache.primary_endpoint
  redis_port                = module.elasticache.port
  redis_tls_enabled         = true
  s3_bucket_name            = module.s3.bucket_name
  s3_bucket_arn             = module.s3.bucket_arn
  s3_endpoint               = var.s3_endpoint
  s3_use_presigned_urls     = var.s3_use_presigned_urls
  app_environment           = var.app_environment
  log_level                 = var.log_level
  prometheus_namespace      = var.prometheus_namespace
  sentry_release            = var.sentry_release
  extra_environment = merge(
    var.extra_environment,
    {
      S3_PUBLIC_BASE_URL          = var.s3_public_base_url
      MEDIA_LOCAL_PUBLIC_BASE_URL = var.media_local_public_base_url
      REDIS_USERNAME              = var.redis_username
    }
  )
  secret_key_mapping = var.secret_env_mapping

  tags = local.common_tags

  depends_on = [
    module.alb
  ]
}
