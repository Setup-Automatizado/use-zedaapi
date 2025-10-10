# Terraform Infrastructure Guide

**Complete guide for deploying WhatsApp API infrastructure on AWS using Terraform**

---

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Environment Configuration](#environment-configuration)
- [Deployment](#deployment)
- [Operations](#operations)
- [Troubleshooting](#troubleshooting)
- [Cost Optimization](#cost-optimization)

---

## 🏗️ Architecture Overview

### Infrastructure Components

```
┌─────────────────────────────────────────────────────────┐
│                      AWS Cloud                          │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │ VPC (10.x.0.0/16)                                 │ │
│  │                                                   │ │
│  │  ┌──────────────┐         ┌──────────────┐      │ │
│  │  │ Public AZ-a  │         │ Public AZ-b  │      │ │
│  │  │              │         │              │      │ │
│  │  │   ┌──────┐   │         │   ┌──────┐   │      │ │
│  │  │   │ ALB  │◄──┼─────────┼───┤ ALB  │   │      │ │
│  │  │   └──┬───┘   │         │   └──┬───┘   │      │ │
│  │  └──────┼───────┘         └──────┼───────┘      │ │
│  │         │                        │               │ │
│  │  ┌──────┴───────┐         ┌──────┴───────┐      │ │
│  │  │ Private AZ-a │         │ Private AZ-b │      │ │
│  │  │              │         │              │      │ │
│  │  │  ┌────────┐  │         │  ┌────────┐  │      │ │
│  │  │  │ECS     │  │         │  │ECS     │  │      │ │
│  │  │  │Task    │  │         │  │Task    │  │      │ │
│  │  │  │(API)   │  │         │  │(API)   │  │      │ │
│  │  │  └──┬─────┘  │         │  └──┬─────┘  │      │ │
│  │  │     │        │         │     │        │      │ │
│  │  │     │        │         │     │        │      │ │
│  │  │     │        │         │     │        │      │ │
│  │  │  ┌──▼────┐   │         │  ┌──▼────┐   │      │ │
│  │  │  │ RDS   │   │         │  │Redis  │   │      │ │
│  │  │  │Postgre│   │         │  │Elasti │   │      │ │
│  │  │  └───────┘   │         │  │Cache  │   │      │ │
│  │  └──────────────┘         └──────────────┘      │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  ┌────────────┐   ┌────────────┐   ┌────────────┐      │
│  │ Secrets    │   │    S3      │   │    ECR     │      │
│  │ Manager    │   │ Media      │   │  Registry  │      │
│  └────────────┘   └────────────┘   └────────────┘      │
└─────────────────────────────────────────────────────────┘
```

### Resource Distribution by Environment

| Resource | Production | Staging | Homolog |
|----------|-----------|---------|---------|
| VPC CIDR | 10.0.0.0/16 | 10.1.0.0/16 | 10.2.0.0/16 |
| NAT Gateway | ✅ Enabled | ✅ Enabled | ❌ Disabled |
| Desired Tasks | 2 | 1 | 1 |
| Max Tasks (Auto-scaling) | 6 | 5 | 2 |
| Container Insights | ✅ Enabled | ❌ Disabled | ❌ Disabled |
| FARGATE_SPOT | ❌ Disabled | ✅ Enabled | ✅ Enabled |
| ECS Exec | ❌ Disabled | ✅ Enabled | ✅ Enabled |
| RDS Multi-AZ | ✅ Enabled | ⚙️ Optional | ❌ Disabled |
| Redis Replicas | 2 | 1 | 0 |
| S3 Force Destroy | ❌ Disabled | ❌ Disabled | ✅ Enabled |

> ⚙️ Optional = configurable via `terraform.tfvars` for that environment.

---

## 📦 Prerequisites

### Required Tools

```bash
# Terraform >= 1.9.0
terraform version

# AWS CLI >= 2.0
aws --version

# jq (for JSON manipulation)
jq --version

# Git (for version control)
git --version
```

### AWS Account Setup

1. **Create IAM User**:
```bash
aws iam create-user --user-name terraform-deployer
aws iam attach-user-policy --user-name terraform-deployer --policy-arn arn:aws:iam::aws:policy/AdministratorAccess
aws iam create-access-key --user-name terraform-deployer
```

2. **Configure AWS CLI**:
```bash
aws configure
# AWS Access Key ID: [from previous step]
# AWS Secret Access Key: [from previous step]
# Default region: us-east-1
# Default output format: json
```

3. **Verify Access**:
```bash
aws sts get-caller-identity
```

---

## 🚀 Initial Setup

### 1. Create S3 Backend (One-Time Setup)

```bash
# Create S3 bucket for Terraform state
aws s3api create-bucket \
  --bucket whatsmeow-terraform-state \
  --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket whatsmeow-terraform-state \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket whatsmeow-terraform-state \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Block public access
aws s3api put-public-access-block \
  --bucket whatsmeow-terraform-state \
  --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true

# Create DynamoDB table for state locking
aws dynamodb create-table \
  --table-name whatsmeow-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

### 2. Generate Credentials

```bash
# Generate strong passwords
export DB_PASSWORD=$(openssl rand -base64 24)
export MINIO_SECRET_KEY=$(openssl rand -base64 24)

# Save to password manager (e.g., 1Password, LastPass)
echo "DB_PASSWORD: $DB_PASSWORD"
echo "MINIO_SECRET_KEY: $MINIO_SECRET_KEY"
```

### 3. (Optional) Create ACM Certificate

```bash
# Request certificate
aws acm request-certificate \
  --domain-name api.example.com \
  --validation-method DNS \
  --region us-east-1

# Get certificate ARN
aws acm list-certificates --region us-east-1

# Follow DNS validation steps in AWS Console
```

---

## ⚙️ Environment Configuration

### Production Setup

```bash
cd terraform/environments/production

# Copy example file
cp terraform.tfvars.example terraform.tfvars

# Edit variables
nano terraform.tfvars
```

**terraform.tfvars** example:
```hcl
aws_region = "us-east-1"

vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]

enable_nat_gateway = false

# certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/..."

api_image = "your-dockerhub-username/whatsmeow-api:latest"

db_user     = "whatsmeow"
db_password = "YOUR_GENERATED_DB_PASSWORD_HERE"

minio_access_key = "minio"
minio_secret_key = "YOUR_GENERATED_MINIO_SECRET_HERE"
```

### Staging/Homolog Setup

Same process as production, but use respective environment directories:
- `terraform/environments/staging/`
- `terraform/environments/homolog/`

**Important**: Each environment uses different VPC CIDR blocks to avoid conflicts:
- Production: `10.0.0.0/16`
- Staging: `10.1.0.0/16`
- Homolog: `10.2.0.0/16`

---

## 🎯 Deployment

### First Deployment

```bash
cd terraform/environments/production

# Initialize Terraform (download providers, configure backend)
terraform init

# Validate configuration
terraform validate

# Preview changes
terraform plan -out=tfplan

# Review plan carefully
# Apply changes
terraform apply tfplan

# Save outputs
terraform output > terraform-outputs.txt
```

### Deployment Order Recommendation

1. **Staging** (test infrastructure)
2. **Homolog** (validate with real data)
3. **Production** (final deployment)

### Subsequent Deployments

```bash
# Pull latest changes
git pull

# Re-initialize (if backend changed)
terraform init -upgrade

# Plan and apply
terraform plan -out=tfplan
terraform apply tfplan
```

---

## 🔧 Operations

### View Infrastructure

```bash
# Show current state
terraform show

# List resources
terraform state list

# Get specific output
terraform output alb_dns_name

# Get all outputs as JSON
terraform output -json
```

### Update ECS Task (Manual)

```bash
# Get cluster name
CLUSTER_NAME=$(terraform output -raw ecs_cluster_name)

# Update service to force new deployment
aws ecs update-service \
  --cluster $CLUSTER_NAME \
  --service whatsmeow-api-service \
  --force-new-deployment
```

### Scale Service

```bash
# Get service name
SERVICE_NAME=$(terraform output -raw ecs_service_name)
CLUSTER_NAME=$(terraform output -raw ecs_cluster_name)

# Update desired count
aws ecs update-service \
  --cluster $CLUSTER_NAME \
  --service $SERVICE_NAME \
  --desired-count 3
```

### Access ECS Container (Debugging)

**Staging/Homolog only** (ECS Exec enabled):

```bash
# List running tasks
CLUSTER_NAME=$(terraform output -raw ecs_cluster_name)
TASK_ID=$(aws ecs list-tasks --cluster $CLUSTER_NAME --query 'taskArns[0]' --output text)

# Execute command in container
aws ecs execute-command \
  --cluster $CLUSTER_NAME \
  --task $TASK_ID \
  --container api \
  --interactive \
  --command "/bin/sh"
```

### View Logs

```bash
# API logs (replace environment name accordingly)
aws logs tail /ecs/production/whatsmeow/api --follow
```

### Update Secrets

```bash
# Get secret ARN
SECRET_ARN=$(terraform output -raw secrets_arn)

# Update secret value
aws secretsmanager update-secret \
  --secret-id $SECRET_ARN \
  --secret-string '{
    "db_user": "whatsmeow",
    "db_password": "new-password-here",
    "postgres_dsn": "postgres://whatsmeow:new-password-here@prod-db.xxxxxx.us-east-1.rds.amazonaws.com:5432/api_core?sslmode=require",
    "wameow_postgres_dsn": "postgres://whatsmeow:new-password-here@prod-db.xxxxxx.us-east-1.rds.amazonaws.com:5432/whatsmeow_store?sslmode=require",
    "redis_password": "optional-auth-token"
  }'

# Force service restart to pick up new secrets
aws ecs update-service \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --service $(terraform output -raw ecs_service_name) \
  --force-new-deployment
```

### Destroy Infrastructure

⚠️ **WARNING**: This will permanently delete all resources!

```bash
# Production: Remove deletion protection first
terraform apply -auto-approve -var="enable_deletion_protection=false"

# Destroy all resources
terraform destroy

# Confirm by typing: yes
```

---

## 🐛 Troubleshooting

### Common Issues

#### 1. **Backend Initialization Failed**

```bash
Error: Failed to get existing workspaces: S3 bucket does not exist
```

**Solution**: Create S3 bucket and DynamoDB table (see [Initial Setup](#initial-setup))

#### 2. **State Lock Error**

```bash
Error: Error acquiring the state lock
```

**Solution**:
```bash
# List locks
aws dynamodb scan --table-name whatsmeow-terraform-locks

# Force unlock (use lock ID from error message)
terraform force-unlock <LOCK_ID>
```

#### 3. **ECS Task Fails to Start**

```bash
# Check service events
aws ecs describe-services \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --services $(terraform output -raw ecs_service_name) \
  --query 'services[0].events[:5]'

# Check task stopped reason
aws ecs describe-tasks \
  --cluster $(terraform output -raw ecs_cluster_name) \
  --tasks $(aws ecs list-tasks --cluster $(terraform output -raw ecs_cluster_name) --query 'taskArns[0]' --output text) \
  --query 'tasks[0].stoppedReason'
```

**Common causes**:
- Secrets not found (check ARN)
- Image pull failed (check ECR permissions)
- Health check failing (check /health endpoint)
- Insufficient resources (check task CPU/memory)

#### 4. **ALB Health Check Failing**

```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn $(terraform output -raw api_target_group_arn)
```

**Common causes**:
- Container not exposing port 8080
- Security group blocking traffic
- Health check path `/health` not responding
- Container taking too long to start (increase `startPeriod`)

#### 5. **RDS Connection Failed**

```bash
# Check RDS status
aws rds describe-db-instances \
  --db-instance-identifier $(terraform output -raw rds_endpoint | cut -d'.' -f1)
```

**Common causes**:
- RDS not in `available` status (wait for maintenance window to finish)
- Security group rules missing (ensure ingress from ECS security group)
- Password rotated without updating secret (see [Update Secrets](#update-secrets))
- `sslmode` misconfigured (`require` is enforced in Terraform templates)

---

## 💰 Cost Optimization

### Monthly Cost Breakdown

#### Production (reference sizing)
- **ECS Fargate** (2 tasks, 2 vCPU/4 GB): ~US$110/month
- **Application Load Balancer**: ~US$25/month
- **RDS PostgreSQL** (db.r6g.large Multi-AZ, 100 GB gp3): ~US$280/month
- **ElastiCache Redis** (cache.r6g.large + replica): ~US$150/month
- **S3 Storage** (100 GB with versioning): ~US$3/month
- **Secrets Manager & KMS**: ~US$2/month
- **CloudWatch Logs & Metrics**: ~US$5/month
- **Estimated total**: **~US$575/month**

#### Staging (single task, SPOT)
- **ECS Fargate** (1 task, SPOT 1 vCPU/2 GB): ~US$15/month
- **ALB**: ~US$20/month
- **RDS PostgreSQL** (db.t4g.medium single AZ, 20 GB): ~US$55/month
- **ElastiCache Redis** (cache.t4g.small + replica): ~US$35/month
- **S3 Storage** (25 GB): ~US$1/month
- **Estimated total**: **~US$126/month**

#### Homolog (single task, SPOT)
- **ECS Fargate** (1 task, SPOT 0.5 vCPU/1 GB): ~US$8/month
- **ALB**: ~US$20/month
- **RDS PostgreSQL** (db.t4g.small, 10 GB): ~US$35/month
- **ElastiCache Redis** (cache.t4g.small, no replica): ~US$20/month
- **S3 Storage** (10 GB): ~US$0.5/month
- **Estimated total**: **~US$84/month**

### Cost Savings Tips

1. **Disable NAT Gateway** (saves $32/month per AZ):
```hcl
enable_nat_gateway = false
```

2. **Use FARGATE_SPOT** (saves 70%):
```hcl
enable_fargate_spot = true
```

3. **Reduce Task Count** (staging/homolog):
```hcl
desired_count = 1
autoscaling_max_capacity = 3
```

4. **Disable Container Insights** (staging/homolog):
```hcl
enable_container_insights = false
```

5. **Right-size Database & Cache**:
   - Use `db_instance_class = "db.t4g.small"` and `redis_node_type = "cache.t4g.small"` for homolog
   - Reduce `db_backup_retention` for non-production (e.g., `3`)

6. **Lifecycle Media Objects**:
   - Configure `s3_lifecycle_rules` to transition older objects to `STANDARD_IA`
   - Enable expiration for presigned artifacts that can be regenerated

### Total Cost Scenarios

| Configuration | Production | Staging | Homolog | Total |
|---------------|-----------|---------|---------|-------|
| **Reference** (current defaults) | $575/mo | $126/mo | $84/mo | **$785/mo** |
| **Optimized** (smaller DB/cache, shorter retention) | $410/mo | $92/mo | $58/mo | **$560/mo** |
| **Minimal** (single staging env, SPOT everywhere) | - | $92/mo | - | **$92/mo** |

---

## 📚 Additional Resources

- [AWS ECS Fargate Documentation](https://docs.aws.amazon.com/ecs/latest/developerguide/AWS_Fargate.html)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS Cost Calculator](https://calculator.aws/)
- [Terraform Best Practices](https://www.terraform-best-practices.com/)

---

**Last Updated**: 2025-10-09
