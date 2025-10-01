# ==================================================
# ECS Service Module - Task Definition + Service
# ==================================================
# Creates:
# - IAM roles for task execution and task
# - Task definition with 4 containers (API, Postgres, Redis, MinIO)
# - ECS service with ALB integration
# - Auto-scaling policies
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

# Allow pulling from ECR and accessing secrets
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
# IAM Role: ECS Task
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

# Allow task to access EFS
resource "aws_iam_role_policy" "task_efs" {
  name = "${var.environment}-whatsmeow-task-efs"
  role = aws_iam_role.task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "elasticfilesystem:ClientMount",
          "elasticfilesystem:ClientWrite"
        ]
        Resource = var.efs_file_system_arn
      }
    ]
  })
}

# ==================================================
# CloudWatch Log Groups
# ==================================================
resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${var.environment}/whatsmeow/api"
  retention_in_days = 7

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "postgres" {
  name              = "/ecs/${var.environment}/whatsmeow/postgres"
  retention_in_days = 7

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "redis" {
  name              = "/ecs/${var.environment}/whatsmeow/redis"
  retention_in_days = 7

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "minio" {
  name              = "/ecs/${var.environment}/whatsmeow/minio"
  retention_in_days = 7

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

  # ==================================================
  # EFS Volume Configuration
  # ==================================================
  volume {
    name = "postgres-data"
    efs_volume_configuration {
      file_system_id     = var.efs_file_system_id
      transit_encryption = "ENABLED"
      authorization_config {
        access_point_id = var.postgres_access_point_id
        iam             = "ENABLED"
      }
    }
  }

  volume {
    name = "redis-data"
    efs_volume_configuration {
      file_system_id     = var.efs_file_system_id
      transit_encryption = "ENABLED"
      authorization_config {
        access_point_id = var.redis_access_point_id
        iam             = "ENABLED"
      }
    }
  }

  volume {
    name = "minio-data"
    efs_volume_configuration {
      file_system_id     = var.efs_file_system_id
      transit_encryption = "ENABLED"
      authorization_config {
        access_point_id = var.minio_access_point_id
        iam             = "ENABLED"
      }
    }
  }

  # ==================================================
  # Container Definitions
  # ==================================================
  container_definitions = jsonencode([
    # API Container
    {
      name      = "api"
      image     = var.api_image
      essential = true
      cpu       = 512
      memory    = 1024

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
          name          = "api"
        }
      ]

      environment = [
        { name = "DB_HOST", value = "localhost" },
        { name = "DB_PORT", value = "5432" },
        { name = "REDIS_HOST", value = "localhost" },
        { name = "REDIS_PORT", value = "6379" },
        { name = "MINIO_ENDPOINT", value = "localhost:9000" },
        { name = "ENVIRONMENT", value = var.environment }
      ]

      secrets = [
        { name = "DB_USER", valueFrom = "${var.secrets_arn}:db_user::" },
        { name = "DB_PASSWORD", valueFrom = "${var.secrets_arn}:db_password::" },
        { name = "MINIO_ACCESS_KEY", valueFrom = "${var.secrets_arn}:minio_access_key::" },
        { name = "MINIO_SECRET_KEY", valueFrom = "${var.secrets_arn}:minio_secret_key::" }
      ]

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:8080/healthz || exit 1"]
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

      dependsOn = [
        { containerName = "postgres", condition = "HEALTHY" },
        { containerName = "redis", condition = "HEALTHY" },
        { containerName = "minio", condition = "HEALTHY" }
      ]
    },

    # Postgres Container
    {
      name      = "postgres"
      image     = "postgres:16-alpine"
      essential = true
      cpu       = 256
      memory    = 512

      portMappings = [
        {
          containerPort = 5432
          protocol      = "tcp"
          name          = "postgres"
        }
      ]

      environment = [
        { name = "POSTGRES_DB", value = "api_core" },
        { name = "PGDATA", value = "/var/lib/postgresql/data/pgdata" }
      ]

      secrets = [
        { name = "POSTGRES_USER", valueFrom = "${var.secrets_arn}:db_user::" },
        { name = "POSTGRES_PASSWORD", valueFrom = "${var.secrets_arn}:db_password::" }
      ]

      mountPoints = [
        {
          sourceVolume  = "postgres-data"
          containerPath = "/var/lib/postgresql/data"
          readOnly      = false
        }
      ]

      healthCheck = {
        command     = ["CMD-SHELL", "pg_isready -U $POSTGRES_USER -d $POSTGRES_DB"]
        interval    = 30
        timeout     = 10
        retries     = 3
        startPeriod = 30
      }

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.postgres.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "postgres"
        }
      }
    },

    # Redis Container
    {
      name      = "redis"
      image     = "redis:7-alpine"
      essential = true
      cpu       = 256
      memory    = 256

      command = ["redis-server", "--appendonly", "yes"]

      portMappings = [
        {
          containerPort = 6379
          protocol      = "tcp"
          name          = "redis"
        }
      ]

      mountPoints = [
        {
          sourceVolume  = "redis-data"
          containerPath = "/data"
          readOnly      = false
        }
      ]

      healthCheck = {
        command     = ["CMD-SHELL", "redis-cli ping || exit 1"]
        interval    = 30
        timeout     = 10
        retries     = 3
        startPeriod = 15
      }

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.redis.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "redis"
        }
      }
    },

    # MinIO Container
    {
      name      = "minio"
      image     = "minio/minio:latest"
      essential = true
      cpu       = 512
      memory    = 1024

      command = ["server", "/data", "--console-address", ":9001"]

      portMappings = [
        {
          containerPort = 9000
          protocol      = "tcp"
          name          = "minio-api"
        },
        {
          containerPort = 9001
          protocol      = "tcp"
          name          = "minio-console"
        }
      ]

      secrets = [
        { name = "MINIO_ROOT_USER", valueFrom = "${var.secrets_arn}:minio_access_key::" },
        { name = "MINIO_ROOT_PASSWORD", valueFrom = "${var.secrets_arn}:minio_secret_key::" }
      ]

      mountPoints = [
        {
          sourceVolume  = "minio-data"
          containerPath = "/data"
          readOnly      = false
        }
      ]

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:9000/minio/health/live || exit 1"]
        interval    = 30
        timeout     = 10
        retries     = 3
        startPeriod = 30
      }

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.minio.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "minio"
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
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.api_target_group_arn
    container_name   = "api"
    container_port   = 8080
  }

  deployment_configuration {
    maximum_percent         = 200
    minimum_healthy_percent = 100
    deployment_circuit_breaker {
      enable   = true
      rollback = true
    }
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
# Auto Scaling Target
# ==================================================
resource "aws_appautoscaling_target" "ecs" {
  count = var.enable_autoscaling ? 1 : 0

  max_capacity       = var.autoscaling_max_capacity
  min_capacity       = var.autoscaling_min_capacity
  resource_id        = "service/${var.cluster_name}/${aws_ecs_service.main.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# ==================================================
# Auto Scaling Policy: CPU
# ==================================================
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

# ==================================================
# Auto Scaling Policy: Memory
# ==================================================
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
