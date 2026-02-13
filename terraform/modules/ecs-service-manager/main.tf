# ==================================================
# ECS Service Manager Module - Next.js Frontend
# ==================================================
# Modulo para deploy do Manager WhatsApp API no ECS
# Utiliza imagem Docker do ECR e conecta ao RDS
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
  container_port = 3000

  # Environment variables for Next.js Manager
  base_environment = {
    # Node/Bun Configuration
    NODE_ENV                   = "production"
    HOSTNAME                   = "0.0.0.0"
    PORT                       = tostring(local.container_port)
    NEXT_TELEMETRY_DISABLED    = "1"

    # Application URLs
    NEXT_PUBLIC_APP_URL        = var.app_url
    BETTER_AUTH_URL            = var.app_url

    # WhatsApp API Backend Connection (internal ALB for server-side)
    WHATSAPP_API_URL           = var.whatsapp_api_url

    # Public API URL (for client-side components and curl examples)
    NEXT_PUBLIC_WHATSAPP_API_URL = var.whatsapp_api_public_url
    NEXT_PUBLIC_API_BASE_URL     = var.whatsapp_api_public_url

    # S3/MinIO Configuration
    S3_ENDPOINT                = var.s3_endpoint
    S3_BUCKET                  = var.s3_bucket
    S3_REGION                  = var.aws_region
    S3_PUBLIC_URL              = var.s3_public_url

    # SMTP Configuration
    SMTP_HOST                  = var.smtp_host
    SMTP_PORT                  = tostring(var.smtp_port)
    SMTP_SECURE                = var.smtp_port == 465 ? "true" : "false"
    SMTP_USER                  = var.smtp_user

    # Email Configuration
    EMAIL_FROM_NAME            = var.email_from_name
    EMAIL_FROM_ADDRESS         = var.email_from_address
    SUPPORT_EMAIL              = var.support_email

    # OAuth Configuration (Client IDs are public)
    GITHUB_CLIENT_ID           = var.github_client_id
    GOOGLE_CLIENT_ID           = var.google_client_id

    # Security Configuration (HTTPS via CloudFront)
    TRUSTED_ORIGINS            = var.app_url
    SECURE_COOKIES             = var.secure_cookies ? "true" : "false"
  }

  environment_definitions = [
    for entry in sort(keys(local.base_environment)) : {
      name  = entry
      value = local.base_environment[entry]
    }
  ]

  # Secrets mapping (from Secrets Manager)
  secret_mapping = {
    DATABASE_URL               = "manager_database_url"
    BETTER_AUTH_SECRET         = "better_auth_secret"
    WHATSAPP_CLIENT_TOKEN      = "client_auth_token"
    WHATSAPP_PARTNER_TOKEN     = "partner_auth_token"
    SMTP_PASSWORD              = "smtp_password"
    S3_ACCESS_KEY              = "s3_access_key"
    S3_SECRET_KEY              = "s3_secret_key"
    GITHUB_CLIENT_SECRET       = "github_client_secret"
    GOOGLE_CLIENT_SECRET       = "google_client_secret"
  }

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
  name = "${var.environment}-manager-task-execution-role"

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
  name = "${var.environment}-manager-task-execution-secrets"
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
  name = "${var.environment}-manager-task-role"

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
  name = "${var.environment}-manager-task-s3"
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

resource "aws_iam_role_policy" "task_ssm" {
  count = var.enable_execute_command ? 1 : 0

  name = "${var.environment}-manager-task-ssm"
  role = aws_iam_role.task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssmmessages:CreateControlChannel",
          "ssmmessages:CreateDataChannel",
          "ssmmessages:OpenControlChannel",
          "ssmmessages:OpenDataChannel"
        ]
        Resource = "*"
      }
    ]
  })
}

# ==================================================
# CloudWatch Log Group
# ==================================================
resource "aws_cloudwatch_log_group" "manager" {
  name              = "/ecs/${var.environment}/manager"
  retention_in_days = 14

  tags = var.tags
}

# ==================================================
# Task Definition: Manager Application
# ==================================================
resource "aws_ecs_task_definition" "main" {
  family                   = "${var.environment}-manager-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([
    {
      name      = "manager"
      image     = var.manager_image
      essential = true
      cpu       = var.task_cpu
      memory    = var.task_memory

      portMappings = [
        {
          containerPort = local.container_port
          protocol      = "tcp"
          name          = "http"
        }
      ]

      environment = local.environment_definitions
      secrets     = local.secret_definitions

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:${local.container_port}/api/health || exit 1"]
        interval    = 30
        timeout     = 10
        retries     = 3
        startPeriod = 60
      }

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.manager.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "manager"
        }
      }
    }
  ])

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-manager-task"
    }
  )
}

# ==================================================
# Task Definition: Database Migration (one-shot)
# ==================================================
resource "aws_ecs_task_definition" "migrate" {
  family                   = "${var.environment}-manager-migrate"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.task_execution.arn
  task_role_arn            = aws_iam_role.task.arn

  container_definitions = jsonencode([
    {
      name      = "migrate"
      image     = var.manager_image
      essential = true
      cpu       = 256
      memory    = 512

      command = ["bun", "prisma", "migrate", "deploy"]

      environment = [
        {
          name  = "NODE_ENV"
          value = "production"
        }
      ]

      secrets = [
        {
          name      = "DATABASE_URL"
          valueFrom = "${var.secrets_arn}:manager_database_url::"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.manager.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "migrate"
        }
      }
    }
  ])

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-manager-migrate"
    }
  )
}

# ==================================================
# ECS Service: Manager
# ==================================================
resource "aws_ecs_service" "main" {
  name            = "${var.environment}-manager-service"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = var.assign_public_ip
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = "manager"
    container_port   = local.container_port
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
      Name = "${var.environment}-manager-service"
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

  name               = "${var.environment}-manager-cpu-scaling"
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

  name               = "${var.environment}-manager-memory-scaling"
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
