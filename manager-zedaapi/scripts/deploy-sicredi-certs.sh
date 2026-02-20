#!/usr/bin/env bash
# =============================================================================
# Deploy Sicredi Certificates to Docker Volume - Manager ZedaAPI
#
# USO:
#   ./scripts/deploy-sicredi-certs.sh
#
# Este script copia os certificados mTLS do Sicredi para o volume Docker
# "zedaapi_sicredi_certs" que e montado nos containers app e worker.
#
# PRE-REQUISITOS:
#   - Docker volume criado: docker volume create zedaapi_sicredi_certs
#   - Certificados na pasta certs/ do projeto
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CERT_DIR="${PROJECT_ROOT}/certs"
VOLUME_NAME="zedaapi_sicredi_certs"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${CYAN}=================================================${NC}"
echo -e "${CYAN}  Manager ZedaAPI - Deploy Certificados Sicredi${NC}"
echo -e "${CYAN}=================================================${NC}"
echo ""

if [ ! -d "$CERT_DIR" ]; then
  echo -e "${RED}Erro: Pasta de certificados nao encontrada: ${CERT_DIR}${NC}"
  echo "Crie a pasta certs/ e adicione os certificados Sicredi."
  exit 1
fi

CERT_COUNT=$(find "$CERT_DIR" -type f \( -name "*.cer" -o -name "*.key" -o -name "*.pem" \) | wc -l)
if [ "$CERT_COUNT" -eq 0 ]; then
  echo -e "${RED}Erro: Nenhum certificado encontrado em ${CERT_DIR}${NC}"
  echo "Adicione os arquivos .cer, .key ou .pem na pasta certs/"
  exit 1
fi

echo -e "${GREEN}[1/3] Verificando volume Docker...${NC}"

if ! docker volume inspect "$VOLUME_NAME" &>/dev/null; then
  echo "  Criando volume: $VOLUME_NAME"
  docker volume create "$VOLUME_NAME"
else
  echo "  Volume ja existe: $VOLUME_NAME"
fi

echo -e "${GREEN}[2/3] Copiando certificados para o volume...${NC}"

docker run --rm \
  -v "$VOLUME_NAME":/certs \
  -v "$CERT_DIR":/source:ro \
  alpine sh -c "
    cp /source/*.cer /certs/ 2>/dev/null || true
    cp /source/*.key /certs/ 2>/dev/null || true
    cp /source/*.pem /certs/ 2>/dev/null || true
    chmod 644 /certs/*.cer 2>/dev/null || true
    chmod 644 /certs/*.pem 2>/dev/null || true
    chmod 600 /certs/*.key 2>/dev/null || true
    chown 1001:1001 /certs/* 2>/dev/null || true
  "

echo -e "${GREEN}[3/3] Verificando conteudo do volume...${NC}"

docker run --rm \
  -v "$VOLUME_NAME":/certs:ro \
  alpine ls -la /certs/

echo ""
echo -e "${GREEN}Certificados implantados com sucesso no volume '${VOLUME_NAME}'!${NC}"
echo ""
echo -e "${YELLOW}Proximos passos:${NC}"
echo "  1. Faca redeploy do stack:"
echo "     docker stack deploy -c docker-compose-swarm.yaml zedaapi"
echo "  2. Registre o webhook:"
echo "     bun run sicredi:webhook:register"
