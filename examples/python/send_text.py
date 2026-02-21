"""
Zé da API — Enviar mensagem de texto
Documentação: https://api.zedaapi.com/docs

Uso: python send_text.py
Requisitos: pip install requests
"""

import os
import sys

import requests

BASE_URL = os.getenv("ZEDAAPI_URL", "https://sua-instancia.zedaapi.com")
CLIENT_TOKEN = os.getenv("ZEDAAPI_TOKEN", "seu-token-aqui")


def send_text(phone: str, message: str) -> dict:
    """Envia uma mensagem de texto via Zé da API."""
    response = requests.post(
        f"{BASE_URL}/send-text",
        headers={
            "Content-Type": "application/json",
            "Client-Token": CLIENT_TOKEN,
        },
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
