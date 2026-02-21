#!/usr/bin/env bash
# =============================================================================
# Zé da API — Conectar instância (obter QR Code)
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

BASE_URL="https://sua-instancia.zedaapi.com"
CLIENT_TOKEN="seu-token-aqui"

# Verificar status da instância
echo "=== Status da Instância ==="
curl -s -X GET "${BASE_URL}/instance-status" \
  -H "Client-Token: ${CLIENT_TOKEN}" | python3 -m json.tool

echo ""

# Obter QR Code para conectar
echo "=== QR Code ==="
curl -s -X GET "${BASE_URL}/qrcode" \
  -H "Client-Token: ${CLIENT_TOKEN}" | python3 -m json.tool
