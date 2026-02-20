#!/usr/bin/env bash
# =============================================================================
# Setup Stripe Webhook - Manager ZedaAPI
#
# USO:
#   # Desenvolvimento local (Stripe CLI forward):
#   ./scripts/setup-stripe-webhook.sh listen
#
#   # Criar webhook endpoint em producao:
#   ./scripts/setup-stripe-webhook.sh create https://seudominio.com.br
#
#   # Listar webhooks existentes:
#   ./scripts/setup-stripe-webhook.sh list
#
#   # Deletar webhook endpoint:
#   ./scripts/setup-stripe-webhook.sh delete we_xxxxxxxx
#
# REQUISITOS:
#   - Stripe CLI instalado: https://docs.stripe.com/stripe-cli
#   - Autenticado: stripe login
# =============================================================================

set -euo pipefail

EVENTS=(
  # Checkout Sessions
  "checkout.session.completed"
  "checkout.session.expired"
  "checkout.session.async_payment_succeeded"
  "checkout.session.async_payment_failed"

  # Payment Intents
  "payment_intent.succeeded"
  "payment_intent.payment_failed"
  "payment_intent.created"

  # Subscriptions
  "customer.subscription.created"
  "customer.subscription.updated"
  "customer.subscription.deleted"

  # Invoices
  "invoice.paid"
  "invoice.payment_failed"
  "invoice.upcoming"

  # Charges
  "charge.refunded"
  "charge.dispute.created"

  # Customer
  "customer.created"
  "customer.updated"
)

EVENTS_CSV=$(IFS=,; echo "${EVENTS[*]}")

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_header() {
  echo -e "\n${CYAN}=================================================${NC}"
  echo -e "${CYAN}  Manager ZedaAPI - Stripe Webhook Setup${NC}"
  echo -e "${CYAN}=================================================${NC}\n"
}

check_stripe_cli() {
  if ! command -v stripe &> /dev/null; then
    echo -e "${RED}Erro: Stripe CLI nao encontrado.${NC}"
    echo -e "Instale: ${YELLOW}brew install stripe/stripe-cli/stripe${NC}"
    echo -e "Ou: ${YELLOW}https://docs.stripe.com/stripe-cli${NC}"
    exit 1
  fi
}

DEV_BASE_URL="http://localhost:3000"

cmd_listen() {
  local endpoint="/api/webhooks/stripe"
  local forward_url="${DEV_BASE_URL}${endpoint}"

  echo -e "${GREEN}Modo: Forward local${NC}"
  echo -e "URL: ${YELLOW}${forward_url}${NC}"
  echo -e "Eventos: ${CYAN}${#EVENTS[@]} tipos${NC}\n"
  echo -e "${YELLOW}Dica: Copie o webhook signing secret (whsec_...) para .env como STRIPE_WEBHOOK_SECRET${NC}\n"

  stripe listen \
    --forward-to "${forward_url}" \
    --events "${EVENTS_CSV}"
}

cmd_create() {
  local base_url="$1"
  local endpoint="/api/webhooks/stripe"
  local full_url="${base_url}${endpoint}"

  echo -e "${GREEN}Modo: Criar endpoint em producao${NC}"
  echo -e "URL: ${YELLOW}${full_url}${NC}"
  echo -e "Eventos: ${CYAN}${#EVENTS[@]} tipos${NC}\n"

  local event_flags=()
  for event in "${EVENTS[@]}"; do
    event_flags+=(--add-event "$event")
  done

  stripe webhook_endpoints create \
    --url "${full_url}" \
    "${event_flags[@]}" \
    --description "Manager ZedaAPI - Webhook principal"

  echo -e "\n${GREEN}Webhook criado com sucesso!${NC}"
  echo -e "${YELLOW}IMPORTANTE: Copie o signing secret (whsec_...) para sua variavel de ambiente STRIPE_WEBHOOK_SECRET${NC}"
}

cmd_list() {
  echo -e "${GREEN}Listando webhooks existentes...${NC}\n"
  stripe webhook_endpoints list
}

cmd_delete() {
  local webhook_id="$1"
  echo -e "${YELLOW}Deletando webhook ${webhook_id}...${NC}"
  stripe webhook_endpoints delete "$webhook_id"
  echo -e "${GREEN}Webhook deletado.${NC}"
}

print_header
check_stripe_cli

case "${1:-help}" in
  listen)
    cmd_listen
    ;;
  create)
    if [ -z "${2:-}" ]; then
      echo -e "${RED}Erro: URL base obrigatoria.${NC}"
      echo -e "Uso: ${YELLOW}$0 create https://seudominio.com.br${NC}"
      exit 1
    fi
    cmd_create "$2"
    ;;
  list)
    cmd_list
    ;;
  delete)
    if [ -z "${2:-}" ]; then
      echo -e "${RED}Erro: ID do webhook obrigatorio.${NC}"
      echo -e "Uso: ${YELLOW}$0 delete we_xxxxxxxx${NC}"
      exit 1
    fi
    cmd_delete "$2"
    ;;
  help|*)
    echo "Uso: $0 <comando> [argumentos]"
    echo ""
    echo "Comandos:"
    echo "  listen                    Forward dev (${DEV_BASE_URL})"
    echo "  create <url-base>         Criar endpoint em producao"
    echo "  list                      Listar webhooks existentes"
    echo "  delete <webhook-id>       Deletar um webhook"
    echo ""
    echo "Eventos configurados (${#EVENTS[@]} tipos):"
    for event in "${EVENTS[@]}"; do
      echo "  - $event"
    done
    ;;
esac
