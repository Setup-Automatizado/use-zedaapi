#!/usr/bin/env bash
# =============================================================================
# Zé da API — Configurar webhook
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

BASE_URL="https://sua-instancia.zedaapi.com"
CLIENT_TOKEN="seu-token-aqui"

# Configurar URLs de webhook para receber eventos
curl -X PUT "${BASE_URL}/webhook" \
  -H "Content-Type: application/json" \
  -H "Client-Token: ${CLIENT_TOKEN}" \
  -d "{
    \"webhookUrl\": \"https://seu-servidor.com/webhook/mensagens\",
    \"messageStatusUrl\": \"https://seu-servidor.com/webhook/status\",
    \"webhookByEvents\": true,
    \"enabled\": true
  }"
