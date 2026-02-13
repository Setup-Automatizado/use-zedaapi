#!/bin/bash
# ==================================================
# Manager Database Migration Script
# ==================================================
# Executa migrations do Prisma no banco de dados
# Uso: ./scripts/migrate.sh [deploy|push|status]
# ==================================================

set -euo pipefail

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar se DATABASE_URL esta definido
if [ -z "${DATABASE_URL:-}" ]; then
    log_error "DATABASE_URL nao definido. Configure a variavel de ambiente."
    exit 1
fi

# Comando padrao
COMMAND="${1:-deploy}"

case "$COMMAND" in
    deploy)
        log_info "Executando migrations em producao..."
        bun prisma migrate deploy
        log_info "Migrations aplicadas com sucesso!"
        ;;
    push)
        log_info "Sincronizando schema com banco (desenvolvimento)..."
        bun prisma db push
        log_info "Schema sincronizado com sucesso!"
        ;;
    status)
        log_info "Verificando status das migrations..."
        bun prisma migrate status
        ;;
    generate)
        log_info "Gerando Prisma Client..."
        bun prisma generate
        log_info "Client gerado com sucesso!"
        ;;
    seed)
        log_info "Executando seed do banco..."
        bun prisma db seed
        log_info "Seed executado com sucesso!"
        ;;
    reset)
        log_warn "ATENCAO: Isso ira resetar o banco de dados!"
        read -p "Tem certeza? (y/N): " confirm
        if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
            bun prisma migrate reset
            log_info "Banco resetado com sucesso!"
        else
            log_info "Operacao cancelada."
        fi
        ;;
    *)
        echo "Uso: $0 [deploy|push|status|generate|seed|reset]"
        echo ""
        echo "Comandos:"
        echo "  deploy   - Aplica migrations pendentes (producao)"
        echo "  push     - Sincroniza schema (desenvolvimento)"
        echo "  status   - Mostra status das migrations"
        echo "  generate - Gera Prisma Client"
        echo "  seed     - Executa seed do banco"
        echo "  reset    - Reseta banco (CUIDADO!)"
        exit 1
        ;;
esac
