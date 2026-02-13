#!/bin/bash
set -e

PROFILE="funnelchat"
REGION="us-east-1"
CLUSTER="production-whatsmeow-cluster"
ACCOUNT="873839854709"
ECR_URL="${ACCOUNT}.dkr.ecr.${REGION}.amazonaws.com"
RDS_HOST="production-whatsmeow-db.ckv4aeiuukxo.us-east-1.rds.amazonaws.com"
DB_USER="whatsmeow"
DB_PASS="80c1f79d907334e75a0403fd79431006bfafdad0634594e13f8194bdb7711a3b"
SUBNET="subnet-0c12bda02ce3cbac7"
SG="sg-02896f9c4cface871"

API_CF="https://d2qqtmk57jbmfv.cloudfront.net"
MANAGER_CF="https://d1zkc6ehods2po.cloudfront.net"

STEP="${1:-help}"

case "$STEP" in

# ============================================
# FASE 3: Criar databases no RDS
# ============================================
create-dbs)
  echo "==> Criando databases no RDS via task temporario..."

  TASK_ARN=$(aws ecs run-task \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --task-definition production-whatsmeow-task \
    --launch-type FARGATE \
    --network-configuration "awsvpcConfiguration={subnets=[$SUBNET],securityGroups=[$SG],assignPublicIp=DISABLED}" \
    --overrides "{\"containerOverrides\":[{\"name\":\"api\",\"command\":[\"sh\",\"-c\",\"apk add --no-cache postgresql16-client && PGPASSWORD=${DB_PASS} psql -h ${RDS_HOST} -U ${DB_USER} -d whatsapp_api -c 'CREATE DATABASE whatsmeow_store;' && PGPASSWORD=${DB_PASS} psql -h ${RDS_HOST} -U ${DB_USER} -d whatsapp_api -c 'CREATE DATABASE manager_db;'\"]}]}" \
    --query 'tasks[0].taskArn' \
    --output text)

  echo "Task iniciada: $TASK_ARN"
  echo ""
  echo "Aguardando task finalizar..."
  aws ecs wait tasks-stopped \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --tasks "$TASK_ARN"

  echo "Task finalizada. Verificando logs..."
  aws logs tail /ecs/production/whatsmeow/api \
    --profile "$PROFILE" \
    --region "$REGION" \
    --since 5m \
    --format short | tail -20

  echo ""
  echo "==> Fase 3 concluida!"
  echo "==> Proximo: ./scripts/deploy-production.sh ecr-login"
  ;;

# ============================================
# FASE 4a: Login ECR
# ============================================
ecr-login)
  echo "==> Login no ECR..."
  aws ecr get-login-password \
    --profile "$PROFILE" \
    --region "$REGION" \
    | docker login --username AWS --password-stdin "$ECR_URL"

  echo "==> Login OK!"
  echo "==> Proximo: ./scripts/deploy-production.sh build-api"
  ;;

# ============================================
# FASE 4b: Build & Push API (Go)
# ============================================
build-api)
  echo "==> Build & Push API (Go)..."
  cd /Users/guilhermejansen/Developer/work/fullstack/whatsapp-api-golang

  # Build for linux/amd64 (ECS Fargate), --load to make image available locally
  docker buildx build \
    -f docker/Dockerfile \
    --target production \
    --platform linux/amd64 \
    --load \
    --build-arg VERSION="$(cat VERSION 2>/dev/null || echo 'latest')" \
    --build-arg COMMIT="$(git rev-parse --short HEAD)" \
    --build-arg BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
    -t "${ECR_URL}/whatsapp-api:latest" \
    .

  docker push "${ECR_URL}/whatsapp-api:latest"

  echo "==> API image pushed!"
  echo "==> Proximo: ./scripts/deploy-production.sh build-manager"
  ;;

# ============================================
# FASE 4c: Build & Push Manager (Next.js)
# ============================================
build-manager)
  echo "==> Build & Push Manager (Next.js)..."

  MANAGER_DIR="/Users/guilhermejansen/Developer/work/fullstack/whatsapp-api-golang/manager-whatsapp-api-golang"

  if [ ! -d "$MANAGER_DIR" ]; then
    echo "ERRO: Diretorio do Manager nao encontrado em $MANAGER_DIR"
    echo "Ajuste o MANAGER_DIR no script"
    exit 1
  fi

  # Build for linux/amd64 (ECS Fargate), --load to make image available locally
  docker buildx build \
    -f "${MANAGER_DIR}/Dockerfile" \
    --platform linux/amd64 \
    --load \
    --build-arg NEXT_PUBLIC_APP_URL="$MANAGER_CF" \
    --build-arg NEXT_PUBLIC_WHATSAPP_API_URL="$API_CF" \
    --build-arg NEXT_PUBLIC_API_BASE_URL="$API_CF" \
    -t "${ECR_URL}/manager-whatsapp-api:latest" \
    "$MANAGER_DIR"

  docker push "${ECR_URL}/manager-whatsapp-api:latest"

  echo "==> Manager image pushed!"
  echo "==> Proximo: ./scripts/deploy-production.sh migrate-manager"
  ;;

# ============================================
# FASE 5: Migration do Manager (Prisma)
# ============================================
migrate-manager)
  echo "==> Rodando migration do Manager..."

  TASK_ARN=$(aws ecs run-task \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --task-definition production-manager-migrate \
    --launch-type FARGATE \
    --network-configuration "awsvpcConfiguration={subnets=[$SUBNET],securityGroups=[$SG],assignPublicIp=DISABLED}" \
    --query 'tasks[0].taskArn' \
    --output text)

  echo "Migration task iniciada: $TASK_ARN"
  echo ""
  echo "Aguardando migration finalizar..."
  aws ecs wait tasks-stopped \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --tasks "$TASK_ARN"

  echo "Migration finalizada. Verificando logs..."
  aws logs tail /ecs/production/manager \
    --profile "$PROFILE" \
    --region "$REGION" \
    --since 5m \
    --format short | tail -20

  echo ""
  echo "==> Fase 5 concluida!"
  echo "==> Proximo: ./scripts/deploy-production.sh deploy-services"
  ;;

# ============================================
# FASE 6: Force Deploy ECS Services
# ============================================
deploy-services)
  echo "==> Force deploy API service..."
  aws ecs update-service \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --service production-whatsmeow-service \
    --force-new-deployment \
    --query 'service.serviceName' \
    --output text

  echo "==> Force deploy Manager service..."
  aws ecs update-service \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --service production-manager-service \
    --force-new-deployment \
    --query 'service.serviceName' \
    --output text

  echo ""
  echo "==> Deploys iniciados! Aguardando estabilizar..."
  aws ecs wait services-stable \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --services production-whatsmeow-service production-manager-service

  echo "==> Servicos estaveis!"
  echo "==> Proximo: ./scripts/deploy-production.sh validate"
  ;;

# ============================================
# FASE 7: Validar
# ============================================
validate)
  echo "==> Validando endpoints..."
  echo ""

  echo "--- API Health ---"
  curl -s -o /dev/null -w "Status: %{http_code}\n" "$API_CF/health" || echo "FALHOU"
  echo ""

  echo "--- Manager Health ---"
  curl -s -o /dev/null -w "Status: %{http_code}\n" "$MANAGER_CF/api/health" || echo "FALHOU"
  echo ""

  echo "--- ECS Services Status ---"
  aws ecs describe-services \
    --profile "$PROFILE" \
    --region "$REGION" \
    --cluster "$CLUSTER" \
    --services production-whatsmeow-service production-manager-service \
    --query 'services[].{Name:serviceName,Status:status,Running:runningCount,Desired:desiredCount}' \
    --output table

  echo ""
  echo "==> URLs finais:"
  echo "    API:     $API_CF"
  echo "    Manager: $MANAGER_CF"
  echo ""
  echo "==> Deploy completo!"
  ;;

# ============================================
# HELP
# ============================================
help|*)
  echo "Deploy Production - WhatsApp API + Manager"
  echo ""
  echo "Uso: ./scripts/deploy-production.sh <step>"
  echo ""
  echo "Steps (executar em ordem):"
  echo "  create-dbs       Fase 3: Criar databases no RDS"
  echo "  ecr-login        Fase 4a: Login no ECR"
  echo "  build-api        Fase 4b: Build & push API (Go)"
  echo "  build-manager    Fase 4c: Build & push Manager (Next.js)"
  echo "  migrate-manager  Fase 5: Rodar Prisma migrations"
  echo "  deploy-services  Fase 6: Force deploy ECS services"
  echo "  validate         Fase 7: Validar endpoints"
  echo ""
  echo "Exemplo:"
  echo "  ./scripts/deploy-production.sh create-dbs"
  ;;

esac
