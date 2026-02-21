#!/usr/bin/env bash
# =============================================================================
# Zé da API — Configurar webhooks
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

# Configuração — substitua pelos seus dados
HOST="https://sua-instancia.zedaapi.com"
INSTANCE_ID="sua-instancia"
INSTANCE_TOKEN="seu-token-aqui"

# URL base com autenticação embutida na rota
BASE_URL="${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}"

# Configurar todas as URLs de webhook de uma vez
curl -X PUT "${BASE_URL}/update-every-webhooks" \
  -H "Content-Type: application/json" \
  -d "{
    \"receivedUrl\": \"https://seu-servidor.com/webhook/mensagens\",
    \"messageStatusUrl\": \"https://seu-servidor.com/webhook/status\",
    \"connectedUrl\": \"https://seu-servidor.com/webhook/conectado\",
    \"disconnectedUrl\": \"https://seu-servidor.com/webhook/desconectado\"
  }"

echo ""
echo "Webhooks configurados com sucesso!"
echo ""
echo "Você também pode configurar cada webhook individualmente:"
echo "  PUT ${BASE_URL}/update-webhook-received"
echo "  PUT ${BASE_URL}/update-webhook-delivery"
echo "  PUT ${BASE_URL}/update-webhook-message-status"
echo "  PUT ${BASE_URL}/update-webhook-connected"
echo "  PUT ${BASE_URL}/update-webhook-disconnected"
echo "  PUT ${BASE_URL}/update-webhook-chat-presence"
echo "  PUT ${BASE_URL}/update-webhook-history-sync"
