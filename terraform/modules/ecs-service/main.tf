# ==================================================
# ECS Service Module - API Task & Service
# ==================================================

terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

locals {
  s3_endpoint = length(var.s3_endpoint) > 0 ? var.s3_endpoint : "https://s3.${var.aws_region}.amazonaws.com"
  gin_mode    = lower(var.app_environment) == "production" ? "release" : "debug"

  base_environment = merge({
    APP_ENV                        = var.app_environment
    ENVIRONMENT                    = var.app_environment
    LOG_LEVEL                      = var.log_level
    GIN_MODE                       = local.gin_mode
    AWS_REGION                     = var.aws_region
    HTTP_ADDR                      = "0.0.0.0:8080"
    HTTP_READ_HEADER_TIMEOUT       = "5s"
    HTTP_READ_TIMEOUT              = "15s"
    HTTP_WRITE_TIMEOUT             = "30s"
    HTTP_IDLE_TIMEOUT              = "120s"
    HTTP_MAX_HEADER_BYTES          = tostring(1048576)
    POSTGRES_MAX_CONNS             = tostring(32)
    DB_HOST                        = var.db_host
    DB_PORT                        = tostring(var.db_port)
    DB_NAME                        = var.db_name_app
    WA_DB_HOST                     = var.db_host
    WA_DB_PORT                     = tostring(var.db_port)
    WA_DB_NAME                     = var.db_name_store
    WAMEOW_LOG_LEVEL               = upper(var.log_level)
    REDIS_ADDR                     = "${var.redis_host}:${var.redis_port}"
    REDIS_USERNAME                 = ""
    REDIS_DB                       = tostring(0)
    REDIS_TLS_ENABLED              = var.redis_tls_enabled ? "true" : "false"
    S3_BUCKET                      = var.s3_bucket_name
    S3_REGION                      = var.aws_region
    S3_ENDPOINT                    = local.s3_endpoint
    S3_USE_SSL                     = "true"
    S3_USE_PRESIGNED_URLS          = var.s3_use_presigned_urls ? "true" : "false"
    S3_PUBLIC_BASE_URL             = ""
    S3_URL_EXPIRATION              = "6d"
    S3_MEDIA_RETENTION             = "720h"
    S3_ACL                         = ""
    MEDIA_LOCAL_STORAGE_PATH       = var.media_local_storage_path
    MEDIA_LOCAL_URL_EXPIRY         = "720h"
    MEDIA_LOCAL_PUBLIC_BASE_URL    = ""
    MEDIA_CLEANUP_INTERVAL         = "168h"
    MEDIA_CLEANUP_BATCH_SIZE       = tostring(200)
    LOCAL_MEDIA_RETENTION          = "720h"
    MEDIA_BUFFER_SIZE              = tostring(500)
    MEDIA_BATCH_SIZE               = tostring(5)
    MEDIA_MAX_RETRIES              = tostring(3)
    MEDIA_POLL_INTERVAL            = "1s"
    MEDIA_DOWNLOAD_TIMEOUT         = "5m"
    MEDIA_UPLOAD_TIMEOUT           = "10m"
    MEDIA_MAX_FILE_SIZE            = tostring(104857600)
    MEDIA_CHUNK_SIZE               = tostring(5242880)
    WEBHOOK_TIMEOUT                = "30s"
    WEBHOOK_MAX_RETRIES            = tostring(3)
    WEBHOOK_DISPATCHER_CONCURRENCY = tostring(8)
    MEDIA_WORKER_CONCURRENCY       = tostring(4)
    PROMETHEUS_NAMESPACE           = var.prometheus_namespace
    EVENT_BUFFER_SIZE              = tostring(1000)
    EVENT_BATCH_SIZE               = tostring(10)
    EVENT_POLL_INTERVAL            = "100ms"
    EVENT_PROCESSING_TIMEOUT       = "30s"
    EVENT_SHUTDOWN_GRACE_PERIOD    = "30s"
    EVENT_MAX_RETRY_ATTEMPTS       = tostring(6)
    EVENT_RETRY_DELAYS             = "0s,10s,30s,2m,5m,15m"
    CB_ENABLED                     = "true"
    CB_MAX_FAILURES                = tostring(5)
    CB_TIMEOUT                     = "60s"
    CB_COOLDOWN                    = "30s"
    DLQ_RETENTION_PERIOD           = "7d"
    DLQ_REPROCESS_ENABLED          = "true"
    TRANSPORT_BUFFER_SIZE          = tostring(100)
    DELIVERED_RETENTION_PERIOD     = "1d"
    CLEANUP_INTERVAL               = "1h"
    EVENTS_DEBUG_RAW_PAYLOAD       = "false"
    EVENTS_DEBUG_DUMP_DIR          = "./tmp/debug-events"
    SENTRY_ENVIRONMENT             = var.app_environment
    SENTRY_RELEASE                 = var.sentry_release
  }, var.extra_environment)

  environment_definitions = [
    for entry in sort(keys(local.base_environment)) : {
      name  = entry
      value = local.base_environment[entry]
    }
  ]

  default_secret_mapping = {
    POSTGRES_DSN        = "postgres_dsn"
    WAMEOW_POSTGRES_DSN = "wameow_postgres_dsn"
    DB_USER             = "db_user"
    DB_PASSWORD         = "db_password"
    REDIS_PASSWORD      = "redis_password"
  }

  secret_mapping = merge(local.default_secret_mapping, var.secret_key_mapping)

  secret_definitions = [
    for env_name, secret_key in local.secret_mapping : {
      name      = env_name
      valueFrom = "${var.secrets_arn}:${secret_key}::"
    }
  ]
}

# ==================================================
# IAM Role: ECS Task Execution
# ==================================================
resource "aws_iam_role" "task_execution" {
  name = "${var.environment}-whatsmeow-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "task_execution_policy" {
  role       = aws_iam_role.task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "task_execution_secrets" {
  name = "${var.environment}-whatsmeow-task-execution-secrets"
  role = aws_iam_role.task_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "kms:Decrypt"
        ]
        Resource = [
          var.secrets_arn,
          "${var.secrets_arn}*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ]
        Resource = "*"
      }
    ]
  })
}

# ==================================================
# IAM Role: ECS Task Runtime
# ==================================================
resource "aws_iam_role" "task" {
  name = "${var.environment}-whatsmeow-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "task_s3" {
  name = "${var.environment}-whatsmeow-task-s3"
  role = aws_iam_role.task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:AbortMultipartUpload",
          "s3:DeleteObject",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject"
        ]
        Resource = [
          var.s3_bucket_arn,
          "${var.s3_bucket_arn}/*"
        ]
      }
    ]
  })
}

# ==================================================
# CloudWatch Log Group
# ==================================================
resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${var.environment}/whatsmeow/api"
  retention_in_days = 14

  tags = var.tags
}

# ==================================================
# Task Definition
# ==================================================
resource "aws_ecs_task_definition" "main" {
  family                   = "${var.environment}-whatsmeow-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([
    {
      name      = "api"
      image     = var.api_image
      essential = true
      cpu       = var.task_cpu
      memory    = var.task_memory

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
          name          = "api"
        }
      ]

      environment = local.environment_definitions
      secrets     = local.secret_definitions

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
        interval    = 30
        timeout     = 10
        retries     = 3
        startPeriod = 60
      }

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.api.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "api"
        }
      }
    }
  ])

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-task"
    }
  )
}

# ==================================================
# ECS Service
# ==================================================
resource "aws_ecs_service" "main" {
  name            = "${var.environment}-whatsmeow-service"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = var.assign_public_ip
  }

  load_balancer {
    target_group_arn = var.api_target_group_arn
    container_name   = "api"
    container_port   = 8080
  }

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  enable_execute_command = var.enable_execute_command

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-service"
    }
  )

  depends_on = [
    aws_iam_role_policy_attachment.task_execution_policy
  ]
}

# ==================================================
# Auto Scaling Target & Policies
# ==================================================
resource "aws_appautoscaling_target" "ecs" {
  count = var.enable_autoscaling ? 1 : 0

  max_capacity       = var.autoscaling_max_capacity
  min_capacity       = var.autoscaling_min_capacity
  resource_id        = "service/${var.cluster_name}/${aws_ecs_service.main.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "cpu" {
  count = var.enable_autoscaling ? 1 : 0

  name               = "${var.environment}-whatsmeow-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = var.autoscaling_cpu_target
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "memory" {
  count = var.enable_autoscaling ? 1 : 0

  name               = "${var.environment}-whatsmeow-memory-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = var.autoscaling_memory_target
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}

# ==================================================
# Data Sources
# ==================================================
data "aws_region" "current" {}
