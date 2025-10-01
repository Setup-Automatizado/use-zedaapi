# ==================================================
# EFS Module - Persistent Storage
# ==================================================
# Creates EFS file system with access points for:
# - Postgres data
# - Redis data
# - MinIO data
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
# EFS File System
# ==================================================
resource "aws_efs_file_system" "main" {
  encrypted        = true
  performance_mode = var.performance_mode
  throughput_mode  = var.throughput_mode

  # Burst mode: Free, up to 100 MiB/s
  # Provisioned: Costs $6/month per MiB/s

  lifecycle_policy {
    transition_to_ia = var.transition_to_ia
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.environment}-whatsmeow-efs"
    }
  )
}

# ==================================================
# EFS Mount Targets (one per AZ)
# ==================================================
resource "aws_efs_mount_target" "main" {
  count = length(var.private_subnet_ids)

  file_system_id  = aws_efs_file_system.main.id
  subnet_id       = var.private_subnet_ids[count.index]
  security_groups = [var.efs_security_group_id]
}

# ==================================================
# Access Point: Postgres Data
# ==================================================
resource "aws_efs_access_point" "postgres" {
  file_system_id = aws_efs_file_system.main.id

  root_directory {
    path = "/postgres"
    creation_info {
      owner_gid   = 999 # postgres user in container
      owner_uid   = 999
      permissions = "0755"
    }
  }

  posix_user {
    gid = 999
    uid = 999
  }

  tags = merge(
    var.tags,
    {
      Name      = "${var.environment}-whatsmeow-efs-postgres"
      Container = "postgres"
    }
  )
}

# ==================================================
# Access Point: Redis Data
# ==================================================
resource "aws_efs_access_point" "redis" {
  file_system_id = aws_efs_file_system.main.id

  root_directory {
    path = "/redis"
    creation_info {
      owner_gid   = 999 # redis user in container
      owner_uid   = 999
      permissions = "0755"
    }
  }

  posix_user {
    gid = 999
    uid = 999
  }

  tags = merge(
    var.tags,
    {
      Name      = "${var.environment}-whatsmeow-efs-redis"
      Container = "redis"
    }
  )
}

# ==================================================
# Access Point: MinIO Data
# ==================================================
resource "aws_efs_access_point" "minio" {
  file_system_id = aws_efs_file_system.main.id

  root_directory {
    path = "/minio"
    creation_info {
      owner_gid   = 1000 # minio user in container
      owner_uid   = 1000
      permissions = "0755"
    }
  }

  posix_user {
    gid = 1000
    uid = 1000
  }

  tags = merge(
    var.tags,
    {
      Name      = "${var.environment}-whatsmeow-efs-minio"
      Container = "minio"
    }
  )
}

# ==================================================
# EFS Backup Policy (optional)
# ==================================================
resource "aws_efs_backup_policy" "main" {
  count = var.enable_backup ? 1 : 0

  file_system_id = aws_efs_file_system.main.id

  backup_policy {
    status = "ENABLED"
  }
}
