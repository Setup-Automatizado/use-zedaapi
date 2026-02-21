#!/usr/bin/env bash
# =============================================================================
# Zé da API — Enviar mensagem de texto
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

# Configuração — substitua pelos seus dados
HOST="https://sua-instancia.zedaapi.com"
INSTANCE_ID="sua-instancia"
INSTANCE_TOKEN="seu-token-aqui"

# URL base com autenticação embutida na rota
BASE_URL="${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}"

curl -X POST "${BASE_URL}/send-text" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"5511999999999\",
    \"message\": \"Olá! Esta é uma mensagem enviada via Zé da API.\"
  }"
