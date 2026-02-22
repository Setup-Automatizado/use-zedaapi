#!/usr/bin/env bash
# =============================================================================
# Zé da API — Conectar instância (obter QR Code)
# Documentação: https://api.zedaapi.com/docs
# =============================================================================

# Configuração — substitua pelos seus dados
HOST="https://sua-instancia.zedaapi.com"
INSTANCE_ID="sua-instancia"
INSTANCE_TOKEN="seu-token-aqui"

# URL base com autenticação embutida na rota
BASE_URL="${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}"

# Verificar status da instância
echo "=== Status da Instância ==="
curl -s -X GET "${BASE_URL}/status" | python3 -m json.tool

echo ""

# Obter QR Code para conectar (retorna base64)
echo "=== QR Code ==="
curl -s -X GET "${BASE_URL}/qr-code" | python3 -m json.tool

echo ""

# Alternativa: obter QR Code como imagem base64
echo "=== QR Code (Imagem Base64) ==="
curl -s -X GET "${BASE_URL}/qr-code/image" | python3 -m json.tool

echo ""

# Alternativa: conectar via código de pareamento (phone pairing code)
echo "=== Código de Pareamento ==="
PHONE="5511999999999"
curl -s -X GET "${BASE_URL}/phone-code/${PHONE}" | python3 -m json.tool
