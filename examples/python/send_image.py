"""
Zé da API — Enviar imagem
Documentação: https://api.zedaapi.com/docs

Uso: python send_image.py
Requisitos: pip install requests
"""

import os
import sys

import requests

BASE_URL = os.getenv("ZEDAAPI_URL", "https://sua-instancia.zedaapi.com")
CLIENT_TOKEN = os.getenv("ZEDAAPI_TOKEN", "seu-token-aqui")


def send_image(phone: str, image_url: str, caption: str | None = None) -> dict:
    """Envia uma imagem via Zé da API."""
    payload: dict = {"phone": phone, "image": image_url}
    if caption:
        payload["caption"] = caption

    response = requests.post(
        f"{BASE_URL}/send-image",
        headers={
            "Content-Type": "application/json",
            "Client-Token": CLIENT_TOKEN,
        },
        json=payload,
        timeout=30,
    )
    response.raise_for_status()
    return response.json()


if __name__ == "__main__":
    try:
        result = send_image(
            "5511999999999",
            "https://exemplo.com/imagem.jpg",
            "Confira esta imagem!",
        )
        print("Imagem enviada com sucesso:")
        print(result)
    except requests.HTTPError as e:
        print(f"Erro HTTP: {e.response.status_code} - {e.response.text}")
        sys.exit(1)
    except requests.RequestException as e:
        print(f"Erro de conexão: {e}")
        sys.exit(1)
