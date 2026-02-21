#!/usr/bin/env bash
# =============================================================================
# Zé da API — Enviar mensagem de texto
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

BASE_URL="https://sua-instancia.zedaapi.com"
CLIENT_TOKEN="seu-token-aqui"

curl -X POST "${BASE_URL}/send-text" \
  -H "Content-Type: application/json" \
  -H "Client-Token: ${CLIENT_TOKEN}" \
  -d "{
    \"phone\": \"5511999999999\",
    \"message\": \"Olá! Esta é uma mensagem enviada via Zé da API.\"
  }"
