#!/bin/bash
# ==================================================
# Manager WhatsApp API - AWS Setup Script
# ==================================================
# Configura recursos AWS necessarios para o Manager
# ==================================================

set -euo pipefail

# Configuracoes
AWS_REGION="${AWS_REGION:-us-east-1}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:-$(aws sts get-caller-identity --query Account --output text)}"
ECR_REPOSITORY="manager-whatsapp-api"
ENVIRONMENT="${ENVIRONMENT:-homolog}"

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

echo "=============================================="
echo "Manager WhatsApp API - AWS Setup"
echo "=============================================="
echo ""
echo "Regiao: $AWS_REGION"
echo "Conta: $AWS_ACCOUNT_ID"
echo "Ambiente: $ENVIRONMENT"
echo ""

# --------------------------------------------------
# Step 1: Criar ECR Repository
# --------------------------------------------------
log_step "1/4 - Criando ECR Repository..."

if aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" &>/dev/null; then
    log_warn "ECR repository '$ECR_REPOSITORY' ja existe"
else
    aws ecr create-repository \
        --repository-name "$ECR_REPOSITORY" \
        --region "$AWS_REGION" \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256
    log_info "ECR repository criado com sucesso"
fi

# --------------------------------------------------
# Step 2: Build e Push da imagem
# --------------------------------------------------
log_step "2/4 - Build e push da imagem Docker..."

ECR_URI="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY"

# Login no ECR
aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

# Build
cd "$(dirname "$0")/.."
docker build -t "$ECR_REPOSITORY:$ENVIRONMENT" .

# Tag e Push
docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$ENVIRONMENT"
docker push "$ECR_URI:$ENVIRONMENT"

log_info "Imagem publicada: $ECR_URI:$ENVIRONMENT"

# --------------------------------------------------
# Step 3: Criar banco de dados no RDS
# --------------------------------------------------
log_step "3/4 - Instrucoes para criar banco de dados..."

echo ""
log_warn "Execute o seguinte comando no PostgreSQL RDS:"
echo ""
cat << 'EOF'
psql -h <RDS_ENDPOINT> -U whatsapp_api_user -d postgres << SQL
CREATE DATABASE manager_db
    WITH OWNER = whatsapp_api_user
    ENCODING = 'UTF8';
SQL
EOF
echo ""

# --------------------------------------------------
# Step 4: Aplicar Terraform
# --------------------------------------------------
log_step "4/4 - Instrucoes para aplicar Terraform..."

echo ""
log_info "Execute os seguintes comandos:"
echo ""
echo "cd terraform/environments/$ENVIRONMENT"
echo "terraform init"
echo "terraform plan -out=tfplan"
echo "terraform apply tfplan"
echo ""

# --------------------------------------------------
# Resumo
# --------------------------------------------------
echo "=============================================="
echo "Setup concluido!"
echo "=============================================="
echo ""
echo "Proximos passos:"
echo "1. Criar banco manager_db no RDS"
echo "2. Configurar secrets no Secrets Manager"
echo "3. Executar terraform apply"
echo "4. Acessar: http://$ENVIRONMENT-whatsmeow-alb-*.amazonaws.com"
echo ""
