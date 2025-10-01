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

  # Remote backend for state management (S3 + DynamoDB)
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
      Repository  = "whatsmeow-private"
    }
  }
}

# ==================================================
# Local Variables
# ==================================================
locals {
  environment = "production"
  common_tags = {
    Environment = "production"
    Project     = "WhatsApp API"
  }
}

# ==================================================
# VPC Module
# ==================================================
module "vpc" {
  source = "../../modules/vpc"

  environment        = local.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  enable_nat_gateway = var.enable_nat_gateway
  enable_flow_logs   = true

  tags = local.common_tags
}

# ==================================================
# Security Groups Module
# ==================================================
module "security_groups" {
  source = "../../modules/security-groups"

  environment          = local.environment
  vpc_id               = module.vpc.vpc_id
  expose_minio_console = false # Production: MinIO console disabled

  tags = local.common_tags
}

# ==================================================
# EFS Module
# ==================================================
module "efs" {
  source = "../../modules/efs"

  environment           = local.environment
  private_subnet_ids    = module.vpc.private_subnet_ids
  efs_security_group_id = module.security_groups.efs_security_group_id
  performance_mode      = "generalPurpose"
  throughput_mode       = "bursting"
  transition_to_ia      = "AFTER_30_DAYS"
  enable_backup         = true # Production: Backups enabled

  tags = local.common_tags
}

# ==================================================
# ALB Module
# ==================================================
module "alb" {
  source = "../../modules/alb"

  environment              = local.environment
  vpc_id                   = module.vpc.vpc_id
  public_subnet_ids        = module.vpc.public_subnet_ids
  alb_security_group_id    = module.security_groups.alb_security_group_id
  certificate_arn          = var.certificate_arn
  expose_minio_console     = false # Production: MinIO console disabled
  enable_deletion_protection = true  # Production: Deletion protection enabled

  tags = local.common_tags
}

# ==================================================
# ECS Cluster Module
# ==================================================
module "ecs_cluster" {
  source = "../../modules/ecs-cluster"

  environment              = local.environment
  enable_container_insights = true  # Production: Container Insights enabled
  enable_fargate_spot      = false # Production: FARGATE_SPOT disabled for stability

  tags = local.common_tags
}

# ==================================================
# Secrets Manager Module
# ==================================================
module "secrets" {
  source = "../../modules/secrets"

  environment             = local.environment
  db_user                 = var.db_user
  db_password             = var.db_password
  minio_access_key        = var.minio_access_key
  minio_secret_key        = var.minio_secret_key
  recovery_window_in_days = 30 # Production: 30 days recovery window

  tags = local.common_tags
}

# ==================================================
# ECS Service Module
# ==================================================
module "ecs_service" {
  source = "../../modules/ecs-service"

  environment               = local.environment
  cluster_id                = module.ecs_cluster.cluster_id
  cluster_name              = module.ecs_cluster.cluster_name
  private_subnet_ids        = module.vpc.private_subnet_ids
  ecs_security_group_id     = module.security_groups.ecs_tasks_security_group_id
  api_target_group_arn      = module.alb.api_target_group_arn
  efs_file_system_id        = module.efs.efs_file_system_id
  efs_file_system_arn       = module.efs.efs_file_system_arn
  postgres_access_point_id  = module.efs.postgres_access_point_id
  redis_access_point_id     = module.efs.redis_access_point_id
  minio_access_point_id     = module.efs.minio_access_point_id
  secrets_arn               = module.secrets.secret_arn
  api_image                 = var.api_image
  task_cpu                  = 1536 # 1.5 vCPU
  task_memory               = 3072 # 3 GB
  desired_count             = 2    # Production: 2 tasks minimum
  enable_execute_command    = false # Production: ECS Exec disabled
  enable_autoscaling        = true
  autoscaling_min_capacity  = 2
  autoscaling_max_capacity  = 10
  autoscaling_cpu_target    = 70
  autoscaling_memory_target = 80

  tags = local.common_tags

  depends_on = [
    module.alb
  ]
}
