#!/bin/bash
# ==================================================
# Manager WhatsApp API - Deploy Script
# ==================================================
# Faz build da imagem Docker e envia para ECR
# Uso: ./scripts/deploy.sh [environment]
# ==================================================

set -euo pipefail

# --------------------------------------------------
# Configuracoes
# --------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
AWS_REGION="${AWS_REGION:-us-east-1}"
ECR_REPOSITORY="manager-whatsapp-api"
ENVIRONMENT="${1:-homolog}"

# URLs por ambiente (OBRIGATORIO para autenticacao funcionar)
# Esta variavel e compilada no build do Next.js, nao pode ser alterada em runtime
# NOTA: Manager tem seu proprio ALB separado da API
if [ "$ENVIRONMENT" = "production" ]; then
    DEFAULT_APP_URL="https://manager.seudominio.com.br"
else
    DEFAULT_APP_URL="http://homolog-manager-alb-936707116.us-east-1.elb.amazonaws.com"
fi

APP_URL="${APP_URL:-$DEFAULT_APP_URL}"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_cmd() { echo -e "${CYAN}[CMD]${NC} $1"; }

# --------------------------------------------------
# Verificar pre-requisitos
# --------------------------------------------------
check_requirements() {
    log_step "Verificando pre-requisitos..."

    # Verificar APP_URL (obrigatorio para autenticacao)
    if [ -z "$APP_URL" ]; then
        log_error "Variavel APP_URL nao definida e ambiente '$ENVIRONMENT' nao reconhecido."
        log_error "Ambientes validos: homolog, production"
        log_error "Ou defina manualmente: APP_URL=http://seu-alb.amazonaws.com ./scripts/deploy.sh"
        exit 1
    fi
    log_info "APP_URL: $APP_URL"

    if ! command -v docker &> /dev/null; then
        log_error "Docker nao esta instalado"
        exit 1
    fi

    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI nao esta instalado"
        exit 1
    fi

    # Verificar se Docker esta rodando
    if ! docker info &> /dev/null; then
        log_error "Docker nao esta rodando"
        exit 1
    fi

    log_info "Pre-requisitos OK"
}

# --------------------------------------------------
# Obter AWS Account ID
# --------------------------------------------------
get_account_id() {
    log_step "Obtendo AWS Account ID..."
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text 2>/dev/null)

    if [ -z "$AWS_ACCOUNT_ID" ]; then
        log_error "Nao foi possivel obter AWS Account ID. Verifique suas credenciais."
        exit 1
    fi

    ECR_URI="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY"
    log_info "Account ID: $AWS_ACCOUNT_ID"
}

# --------------------------------------------------
# Login no ECR
# --------------------------------------------------
ecr_login() {
    log_step "Fazendo login no ECR..."
    log_cmd "aws ecr get-login-password --region $AWS_REGION | docker login ..."

    aws ecr get-login-password --region "$AWS_REGION" | \
        docker login --username AWS --password-stdin \
        "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"

    log_info "Login no ECR realizado com sucesso"
}

# --------------------------------------------------
# Build da imagem
# --------------------------------------------------
build_image() {
    log_step "Construindo imagem Docker..."

    cd "$PROJECT_DIR"

    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    VERSION="$ENVIRONMENT"
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    log_cmd "docker build --platform linux/amd64 --build-arg NEXT_PUBLIC_APP_URL=$APP_URL --build-arg BUILD_TIME=$BUILD_TIME --build-arg VERSION=$VERSION --build-arg COMMIT=$COMMIT -t $ECR_REPOSITORY:$ENVIRONMENT ."

    docker build \
        --platform linux/amd64 \
        --build-arg NEXT_PUBLIC_APP_URL="$APP_URL" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg VERSION="$VERSION" \
        --build-arg COMMIT="$COMMIT" \
        -t "$ECR_REPOSITORY:$ENVIRONMENT" \
        .

    log_info "Build concluido com sucesso"
}

# --------------------------------------------------
# Tag e Push da imagem
# --------------------------------------------------
push_image() {
    log_step "Enviando imagem para ECR..."

    # Tag com ambiente
    log_cmd "docker tag $ECR_REPOSITORY:$ENVIRONMENT $ECR_URI:$ENVIRONMENT"
    docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$ENVIRONMENT"

    # Tag com commit SHA
    COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    log_cmd "docker tag $ECR_REPOSITORY:$ENVIRONMENT $ECR_URI:$COMMIT"
    docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$COMMIT"

    # Push das duas tags
    log_cmd "docker push $ECR_URI:$ENVIRONMENT"
    docker push "$ECR_URI:$ENVIRONMENT"

    log_cmd "docker push $ECR_URI:$COMMIT"
    docker push "$ECR_URI:$COMMIT"

    log_info "Imagem enviada com sucesso"
}

# --------------------------------------------------
# Atualizar servico ECS (opcional)
# --------------------------------------------------
update_ecs_service() {
    ECS_CLUSTER="${ENVIRONMENT}-whatsmeow-cluster"
    ECS_SERVICE="${ENVIRONMENT}-manager-service"

    log_step "Verificando se servico ECS existe..."

    if aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$AWS_REGION" &>/dev/null; then
        log_step "Atualizando servico ECS..."
        log_cmd "aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --force-new-deployment"

        aws ecs update-service \
            --cluster "$ECS_CLUSTER" \
            --service "$ECS_SERVICE" \
            --force-new-deployment \
            --region "$AWS_REGION" > /dev/null

        log_info "Servico ECS atualizado. Aguardando estabilizacao..."

        if [ "${WAIT_FOR_STABLE:-false}" = "true" ]; then
            aws ecs wait services-stable \
                --cluster "$ECS_CLUSTER" \
                --services "$ECS_SERVICE" \
                --region "$AWS_REGION"
            log_info "Servico estavel"
        else
            log_warn "Use WAIT_FOR_STABLE=true para aguardar estabilizacao"
        fi
    else
        log_warn "Servico ECS nao encontrado. Execute terraform apply primeiro."
    fi
}

# --------------------------------------------------
# Main
# --------------------------------------------------
main() {
    echo ""
    echo "=============================================="
    echo "   Manager WhatsApp API - Deploy"
    echo "=============================================="
    echo ""
    echo "Ambiente:   $ENVIRONMENT"
    echo "Regiao:     $AWS_REGION"
    echo "Diretorio:  $PROJECT_DIR"
    echo ""

    check_requirements
    get_account_id
    ecr_login
    build_image
    push_image
    update_ecs_service

    echo ""
    echo "=============================================="
    echo "   Deploy concluido com sucesso!"
    echo "=============================================="
    echo ""
    echo "Imagem: $ECR_URI:$ENVIRONMENT"
    echo ""
    echo "Proximos passos:"
    echo "  1. Verifique o status no ECS Console"
    echo "  2. Acesse: http://${ENVIRONMENT}-whatsmeow-alb-*.amazonaws.com"
    echo ""
}

main
