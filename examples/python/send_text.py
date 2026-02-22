"""
Zé da API — Enviar mensagem de texto
Documentação: https://api.zedaapi.com/docs

Uso: python send_text.py
Requisitos: pip install requests
"""

import os
import sys

import requests

# Configuração — substitua pelos seus dados
HOST = os.getenv("ZEDAAPI_HOST", "https://sua-instancia.zedaapi.com")
INSTANCE_ID = os.getenv("ZEDAAPI_INSTANCE_ID", "sua-instancia")
INSTANCE_TOKEN = os.getenv("ZEDAAPI_INSTANCE_TOKEN", "seu-token-aqui")

# URL base com autenticação embutida na rota
BASE_URL = f"{HOST}/instances/{INSTANCE_ID}/token/{INSTANCE_TOKEN}"


def send_text(phone: str, message: str) -> dict:
    """Envia uma mensagem de texto via Zé da API."""
    response = requests.post(
        f"{BASE_URL}/send-text",
        headers={"Content-Type": "application/json"},
        json={"phone": phone, "message": message},
        timeout=30,
    )
    response.raise_for_status()
    return response.json()


if __name__ == "__main__":
    try:
        result = send_text("5511999999999", "Olá! Mensagem enviada via Python.")
        print("Mensagem enviada com sucesso:")
        print(result)
    except requests.HTTPError as e:
        print(f"Erro HTTP: {e.response.status_code} - {e.response.text}")
        sys.exit(1)
    except requests.RequestException as e:
        print(f"Erro de conexão: {e}")
        sys.exit(1)
