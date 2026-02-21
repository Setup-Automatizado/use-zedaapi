#!/usr/bin/env bash
# =============================================================================
# Zé da API — Enviar documento
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

BASE_URL="https://sua-instancia.zedaapi.com"
CLIENT_TOKEN="seu-token-aqui"

curl -X POST "${BASE_URL}/send-document" \
  -H "Content-Type: application/json" \
  -H "Client-Token: ${CLIENT_TOKEN}" \
  -d "{
    \"phone\": \"5511999999999\",
    \"document\": \"https://exemplo.com/relatorio.pdf\",
    \"fileName\": \"relatorio.pdf\",
    \"caption\": \"Segue o relatório solicitado.\"
  }"
