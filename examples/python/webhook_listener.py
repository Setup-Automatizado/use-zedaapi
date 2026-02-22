"""
Zé da API — Listener de Webhooks
Documentação: https://api.zedaapi.com/docs

Servidor Flask simples que recebe eventos do Zé da API via webhook.

Uso: python webhook_listener.py
Requisitos: pip install flask
"""

import json
import os
from datetime import datetime, timezone

from flask import Flask, jsonify, request

app = Flask(__name__)

PORT = int(os.getenv("PORT", "3000"))


@app.route("/health", methods=["GET"])
def health():
    """Health check."""
    return jsonify({"status": "ok"})


@app.route("/webhook", methods=["POST"])
def webhook():
    """Recebe eventos do Zé da API."""
    event = request.get_json(silent=True)
    if event is None:
        return jsonify({"error": "Payload inválido"}), 400

    event_type = event.get("event", "unknown")
    timestamp = datetime.now(timezone.utc).isoformat()

    print(f"[{timestamp}] Evento recebido: {event_type}")
    print(json.dumps(event, indent=2, ensure_ascii=False))
    print("---")

    # Processar diferentes tipos de evento
    if event_type == "messages.upsert":
        print("Nova mensagem recebida!")
    elif event_type == "messages.update":
        print("Status de mensagem atualizado!")
    elif event_type == "connection.update":
        print("Status da conexão alterado!")
    else:
        print(f"Evento não tratado: {event_type}")

    # Sempre retornar 200 rapidamente
    return jsonify({"received": True})


if __name__ == "__main__":
    print(f"Webhook listener rodando em http://localhost:{PORT}/webhook")
    print("Aguardando eventos...")
    print()
    print("Configure o webhook no Zé da API apontando para:")
    print(f"  http://SEU-IP:{PORT}/webhook")
    print()
    app.run(host="0.0.0.0", port=PORT, debug=True)
