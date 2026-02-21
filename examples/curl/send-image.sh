#!/usr/bin/env bash
# =============================================================================
# Zé da API — Enviar imagem
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

BASE_URL="https://sua-instancia.zedaapi.com"
CLIENT_TOKEN="seu-token-aqui"

curl -X POST "${BASE_URL}/send-image" \
  -H "Content-Type: application/json" \
  -H "Client-Token: ${CLIENT_TOKEN}" \
  -d "{
    \"phone\": \"5511999999999\",
    \"image\": \"https://exemplo.com/imagem.jpg\",
    \"caption\": \"Confira esta imagem!\"
  }"
