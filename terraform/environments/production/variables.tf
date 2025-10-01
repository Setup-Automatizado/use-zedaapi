# ==================================================
# Production Environment - Variables
# ==================================================

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway (adds ~$32/month per AZ)"
  type        = bool
  default     = false
}

variable "certificate_arn" {
  description = "ACM certificate ARN for HTTPS (optional)"
  type        = string
  default     = null
}

variable "api_image" {
  description = "Docker image for API container"
  type        = string
  default     = "whatsmeow-api:latest"
}

variable "db_user" {
  description = "Database username"
  type        = string
  default     = "whatsmeow"
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "minio_access_key" {
  description = "MinIO access key"
  type        = string
  default     = "minio"
  sensitive   = true
}

variable "minio_secret_key" {
  description = "MinIO secret key"
  type        = string
  sensitive   = true
}
