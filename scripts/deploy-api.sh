#!/bin/bash
# ==================================================
# WhatsApp API Golang - Deploy Script
# ==================================================
# Faz build da imagem Docker e envia para ECR
# Uso: ./scripts/deploy-api.sh [environment]
# ==================================================

set -euo pipefail

# --------------------------------------------------
# Configuracoes
# --------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
AWS_REGION="${AWS_REGION:-us-east-1}"
ECR_REPOSITORY="whatsapp-api"
ENVIRONMENT="${1:-homolog}"
DOCKERFILE="docker/Dockerfile"

# URLs por ambiente
if [ "$ENVIRONMENT" = "production" ]; then
    API_URL="https://api.seudominio.com.br"
else
    API_URL="http://homolog-whatsmeow-alb-731186848.us-east-1.elb.amazonaws.com"
fi

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

    # Verificar se Dockerfile existe
    if [ ! -f "$PROJECT_DIR/$DOCKERFILE" ]; then
        log_error "Dockerfile nao encontrado em $PROJECT_DIR/$DOCKERFILE"
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
# Criar repositorio ECR se nao existir
# --------------------------------------------------
ensure_ecr_repository() {
    log_step "Verificando repositorio ECR..."

    if ! aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" &>/dev/null; then
        log_warn "Repositorio $ECR_REPOSITORY nao existe. Criando..."
        aws ecr create-repository \
            --repository-name "$ECR_REPOSITORY" \
            --region "$AWS_REGION" \
            --image-scanning-configuration scanOnPush=true \
            --encryption-configuration encryptionType=AES256 > /dev/null
        log_info "Repositorio criado com sucesso"
    else
        log_info "Repositorio $ECR_REPOSITORY ja existe"
    fi
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

    log_cmd "docker build --platform linux/amd64 -f $DOCKERFILE --build-arg BUILD_TIME=$BUILD_TIME --build-arg VERSION=$VERSION --build-arg COMMIT=$COMMIT --target production -t $ECR_REPOSITORY:$ENVIRONMENT --load ."

    docker build \
        --platform linux/amd64 \
        -f "$DOCKERFILE" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg VERSION="$VERSION" \
        --build-arg COMMIT="$COMMIT" \
        --target production \
        -t "$ECR_REPOSITORY:$ENVIRONMENT" \
        --load \
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
# Atualizar servico ECS
# --------------------------------------------------
update_ecs_service() {
    ECS_CLUSTER="${ENVIRONMENT}-whatsmeow-cluster"
    ECS_SERVICE="${ENVIRONMENT}-whatsmeow-service"

    log_step "Verificando se servico ECS existe..."

    if aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$AWS_REGION" 2>/dev/null | grep -q "\"status\": \"ACTIVE\""; then
        log_step "Atualizando servico ECS..."
        log_cmd "aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --force-new-deployment"

        aws ecs update-service \
            --cluster "$ECS_CLUSTER" \
            --service "$ECS_SERVICE" \
            --force-new-deployment \
            --region "$AWS_REGION" > /dev/null

        log_info "Servico ECS atualizado. Aguardando estabilizacao..."

        if [ "${WAIT_FOR_STABLE:-false}" = "true" ]; then
            log_step "Aguardando servico estabilizar..."
            aws ecs wait services-stable \
                --cluster "$ECS_CLUSTER" \
                --services "$ECS_SERVICE" \
                --region "$AWS_REGION"
            log_info "Servico estavel"
        else
            log_warn "Use WAIT_FOR_STABLE=true para aguardar estabilizacao"
        fi
    else
        log_warn "Servico ECS '$ECS_SERVICE' nao encontrado no cluster '$ECS_CLUSTER'."
        log_warn "Verifique se o terraform foi aplicado ou se o nome do servico esta correto."
    fi
}

# --------------------------------------------------
# Main
# --------------------------------------------------
main() {
    echo ""
    echo "=============================================="
    echo "   WhatsApp API Golang - Deploy"
    echo "=============================================="
    echo ""
    echo "Ambiente:   $ENVIRONMENT"
    echo "Regiao:     $AWS_REGION"
    echo "Diretorio:  $PROJECT_DIR"
    echo "Dockerfile: $DOCKERFILE"
    echo "API URL:    $API_URL"
    echo ""

    check_requirements
    get_account_id
    ecr_login
    ensure_ecr_repository
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
    echo "  2. Monitore os logs no CloudWatch"
    echo "  3. Teste o health check: GET /health"
    echo ""
}

main
