#!/bin/bash
# ==================================================
# WhatsApp API - Deploy All Services
# ==================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
AWS_REGION="${AWS_REGION:-us-east-1}"
ENVIRONMENT="${1:-homolog}"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_cmd() { echo -e "${CYAN}[CMD]${NC} $1"; }
log_header() { echo -e "${MAGENTA}$1${NC}"; }

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

    if ! docker info &> /dev/null; then
        log_error "Docker nao esta rodando"
        exit 1
    fi

    log_info "Pre-requisitos OK"
}

get_account_id() {
    log_step "Obtendo AWS Account ID..."
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text 2>/dev/null)

    if [ -z "$AWS_ACCOUNT_ID" ]; then
        log_error "Nao foi possivel obter AWS Account ID. Verifique suas credenciais."
        exit 1
    fi

    ECR_BASE="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"
    log_info "Account ID: $AWS_ACCOUNT_ID"
}

ecr_login() {
    log_step "Fazendo login no ECR..."

    aws ecr get-login-password --region "$AWS_REGION" | \
        docker login --username AWS --password-stdin "$ECR_BASE"

    log_info "Login no ECR realizado com sucesso"
}

ensure_ecr_repository() {
    local repo_name="$1"
    log_step "Verificando repositorio ECR: $repo_name..."

    if ! aws ecr describe-repositories --repository-names "$repo_name" --region "$AWS_REGION" &>/dev/null; then
        log_warn "Repositorio $repo_name nao existe. Criando..."
        aws ecr create-repository \
            --repository-name "$repo_name" \
            --region "$AWS_REGION" \
            --image-scanning-configuration scanOnPush=true \
            --encryption-configuration encryptionType=AES256 > /dev/null
        log_info "Repositorio criado com sucesso"
    else
        log_info "Repositorio $repo_name ja existe"
    fi
}

# --------------------------------------------------
# Build, Tag e Push - API
# --------------------------------------------------
deploy_api() {
    log_header ""
    log_header "=============================================="
    log_header "   Deploying: WhatsApp API (Backend)"
    log_header "=============================================="
    log_header ""

    local ECR_REPOSITORY="whatsapp-api"
    local ECR_URI="$ECR_BASE/$ECR_REPOSITORY"
    local DOCKERFILE="docker/Dockerfile"

    if [ ! -f "$PROJECT_DIR/$DOCKERFILE" ]; then
        log_error "Dockerfile nao encontrado em $PROJECT_DIR/$DOCKERFILE"
        return 1
    fi

    ensure_ecr_repository "$ECR_REPOSITORY"

    cd "$PROJECT_DIR"

    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    VERSION="${VERSION:-$ENVIRONMENT}"
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    log_step "Building API image..."
    log_cmd "docker build --platform linux/amd64 -f $DOCKERFILE --build-arg VERSION=$VERSION --build-arg COMMIT=$COMMIT --build-arg BUILD_TIME=$BUILD_TIME -t $ECR_REPOSITORY:$ENVIRONMENT ."

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

    log_step "Pushing API image..."
    docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$ENVIRONMENT"
    docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$COMMIT"
    docker push "$ECR_URI:$ENVIRONMENT"
    docker push "$ECR_URI:$COMMIT"

    log_info "API image pushed: $ECR_URI:$ENVIRONMENT"

    # Update ECS service
    local ECS_CLUSTER="${ENVIRONMENT}-whatsmeow-cluster"
    local ECS_SERVICE="${ENVIRONMENT}-whatsmeow-service"

    if aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$AWS_REGION" 2>/dev/null | grep -q "\"status\": \"ACTIVE\""; then
        log_step "Updating ECS service: $ECS_SERVICE..."
        aws ecs update-service \
            --cluster "$ECS_CLUSTER" \
            --service "$ECS_SERVICE" \
            --force-new-deployment \
            --region "$AWS_REGION" > /dev/null
        log_info "API ECS service updated"
    else
        log_warn "API ECS service not found: $ECS_SERVICE"
    fi
}

# --------------------------------------------------
# Build, Tag e Push - Manager
# --------------------------------------------------
deploy_manager() {
    log_header ""
    log_header "=============================================="
    log_header "   Deploying: Manager (Frontend)"
    log_header "=============================================="
    log_header ""

    local ECR_REPOSITORY="manager-whatsapp-api"
    local ECR_URI="$ECR_BASE/$ECR_REPOSITORY"
    local MANAGER_DIR="$PROJECT_DIR/manager-whatsapp-api-golang"

    if [ ! -d "$MANAGER_DIR" ]; then
        log_error "Manager directory not found: $MANAGER_DIR"
        return 1
    fi

    if [ ! -f "$MANAGER_DIR/Dockerfile" ]; then
        log_error "Dockerfile nao encontrado em $MANAGER_DIR/Dockerfile"
        return 1
    fi

    ensure_ecr_repository "$ECR_REPOSITORY"

    cd "$MANAGER_DIR"

    if [ "$ENVIRONMENT" = "production" ]; then
        DEFAULT_APP_URL="https://manager.funnelchat.com"
    else
        DEFAULT_APP_URL="http://homolog-manager-alb-936707116.us-east-1.elb.amazonaws.com"
    fi
    APP_URL="${APP_URL:-$DEFAULT_APP_URL}"

    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    VERSION="${VERSION:-$ENVIRONMENT}"
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    log_step "Building Manager image..."
    log_cmd "docker build --platform linux/amd64 --build-arg NEXT_PUBLIC_APP_URL=$APP_URL -t $ECR_REPOSITORY:$ENVIRONMENT ."

    docker build \
        --platform linux/amd64 \
        --load \
        --build-arg NEXT_PUBLIC_APP_URL="$APP_URL" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg VERSION="$VERSION" \
        --build-arg COMMIT="$COMMIT" \
        -t "$ECR_REPOSITORY:$ENVIRONMENT" \
        .

    log_step "Pushing Manager image..."
    docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$ENVIRONMENT"
    docker tag "$ECR_REPOSITORY:$ENVIRONMENT" "$ECR_URI:$COMMIT"
    docker push "$ECR_URI:$ENVIRONMENT"
    docker push "$ECR_URI:$COMMIT"

    log_info "Manager image pushed: $ECR_URI:$ENVIRONMENT"

    # Update ECS service
    local ECS_CLUSTER="${ENVIRONMENT}-whatsmeow-cluster"
    local ECS_SERVICE="${ENVIRONMENT}-manager-service"

    if aws ecs describe-services --cluster "$ECS_CLUSTER" --services "$ECS_SERVICE" --region "$AWS_REGION" 2>/dev/null | grep -q "\"status\": \"ACTIVE\""; then
        log_step "Updating ECS service: $ECS_SERVICE..."
        aws ecs update-service \
            --cluster "$ECS_CLUSTER" \
            --service "$ECS_SERVICE" \
            --force-new-deployment \
            --region "$AWS_REGION" > /dev/null
        log_info "Manager ECS service updated"
    else
        log_warn "Manager ECS service not found: $ECS_SERVICE"
    fi

    cd "$PROJECT_DIR"
}

# --------------------------------------------------
# Wait for services to stabilize
# --------------------------------------------------
wait_for_services() {
    if [ "${WAIT_FOR_STABLE:-false}" != "true" ]; then
        log_warn "Use WAIT_FOR_STABLE=true to wait for service stabilization"
        return
    fi

    local ECS_CLUSTER="${ENVIRONMENT}-whatsmeow-cluster"

    log_step "Waiting for services to stabilize..."

    # Wait for API
    if aws ecs describe-services --cluster "$ECS_CLUSTER" --services "${ENVIRONMENT}-whatsmeow-service" --region "$AWS_REGION" 2>/dev/null | grep -q "\"status\": \"ACTIVE\""; then
        log_info "Waiting for API service..."
        aws ecs wait services-stable \
            --cluster "$ECS_CLUSTER" \
            --services "${ENVIRONMENT}-whatsmeow-service" \
            --region "$AWS_REGION" || true
    fi

    # Wait for Manager
    if aws ecs describe-services --cluster "$ECS_CLUSTER" --services "${ENVIRONMENT}-manager-service" --region "$AWS_REGION" 2>/dev/null | grep -q "\"status\": \"ACTIVE\""; then
        log_info "Waiting for Manager service..."
        aws ecs wait services-stable \
            --cluster "$ECS_CLUSTER" \
            --services "${ENVIRONMENT}-manager-service" \
            --region "$AWS_REGION" || true
    fi

    log_info "Services stabilized"
}

# --------------------------------------------------
# Main
# --------------------------------------------------
main() {
    echo ""
    echo "=============================================="
    echo "   WhatsApp API - Deploy All Services"
    echo "=============================================="
    echo ""
    echo "Environment: $ENVIRONMENT"
    echo "Region:      $AWS_REGION"
    echo "Directory:   $PROJECT_DIR"
    echo ""

    check_requirements
    get_account_id
    ecr_login

    # Deploy both services
    deploy_api
    deploy_manager

    # Wait for stabilization if requested
    wait_for_services

    echo ""
    echo "=============================================="
    echo "   Deploy completed successfully!"
    echo "=============================================="
    echo ""
    echo "Services deployed:"
    echo "  - API:     $ECR_BASE/whatsapp-api:$ENVIRONMENT"
    echo "  - Manager: $ECR_BASE/manager-whatsapp-api:$ENVIRONMENT"
    echo ""
    echo "Next steps:"
    echo "  1. Check ECS Console for service status"
    echo "  2. Monitor CloudWatch logs"
    echo "  3. Test API health: GET /health"
    echo ""
}

main
