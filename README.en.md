# Ze da API - Support Hub

[![Version](https://img.shields.io/badge/version-v3.10.0-blue.svg)](https://github.com/Setup-Automatizado/use-zedaapi/releases)
[![Status](https://img.shields.io/badge/status-active-brightgreen.svg)](https://status.zedaapi.com)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539.svg)](openapi/openapi-latest.json)
[![n8n](https://img.shields.io/badge/n8n-community%20node-FF6D5A.svg)](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)

**English** | **[Español](README.es.md)** | **[Português](README.md)**

Official support hub for **Ze da API** subscribers — the most complete WhatsApp API from Brazil.

Here you'll find the OpenAPI specification, Postman collection, n8n node, code examples, and the full changelog.

---

## Quick Links

| Resource | Link |
|----------|------|
| API Documentation | [api.zedaapi.com/docs](https://api.zedaapi.com/docs) |
| Dashboard | [zedaapi.com](https://zedaapi.com) |
| OpenAPI JSON | [openapi-latest.json](openapi/openapi-latest.json) |
| OpenAPI YAML | [openapi-latest.yaml](openapi/openapi-latest.yaml) |
| Postman Collection | [zedaapi-latest.postman.json](postman/zedaapi-latest.postman.json) |
| n8n Community Node | [n8n-nodes-zedaapi](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Status | [status.zedaapi.com](https://status.zedaapi.com) |

---

## Quick Start

Send your first message in 30 seconds:

```bash
# Replace with your actual credentials
HOST="https://your-instance.zedaapi.com"
INSTANCE_ID="your-instance"
INSTANCE_TOKEN="your-token-here"

curl -X POST "${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}/send-text" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "message": "Hello! Message sent via Ze da API."
  }'
```

> Authentication is done via **Instance ID** and **Instance Token** directly in the URL. Check the dashboard for your credentials.

---

## Postman Collection

Import the full collection into Postman to test all API routes:

1. Download [`zedaapi-latest.postman.json`](postman/zedaapi-latest.postman.json)
2. In Postman, click **Import** and select the file
3. Set up the collection variables:
   - `baseUrl`: Server URL (e.g., `https://your-instance.zedaapi.com`)
   - `instanceId`: Your instance ID
   - `instanceToken`: Your instance token

The collection includes all routes organized by category: System, Instance, Messages, Webhooks, Contacts, Groups, Media, and more.

---

## OpenAPI Spec

The OpenAPI 3.1 specification is available in two formats:

- **JSON**: [`openapi/openapi-latest.json`](openapi/openapi-latest.json)
- **YAML**: [`openapi/openapi-latest.yaml`](openapi/openapi-latest.yaml)

You can import it into any OpenAPI-compatible tool (Swagger UI, Insomnia, Stoplight, etc).

Previous versions are available as `openapi-vX.Y.Z.json`.

---

## n8n Community Node

Integrate Ze da API directly into your n8n workflows with the official node:

**[n8n-nodes-zedaapi](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)** — Send messages, manage groups, communities, newsletters, configure webhooks, and more, all from n8n.

### Installing in n8n

1. Go to **Settings > Community Nodes**
2. Click **Install a community node**
3. Type `n8n-nodes-zedaapi` and install

The node supports all API operations: sending messages (text, image, document, audio, video), instance management, webhooks, groups, communities, and newsletters.

---

## Code Examples

### cURL

Ready-to-use shell scripts:

| Script | Description |
|--------|-------------|
| [`send-text.sh`](examples/curl/send-text.sh) | Send a text message |
| [`send-image.sh`](examples/curl/send-image.sh) | Send an image |
| [`send-document.sh`](examples/curl/send-document.sh) | Send a document |
| [`instance-connect.sh`](examples/curl/instance-connect.sh) | Connect instance (QR Code) |
| [`webhook-setup.sh`](examples/curl/webhook-setup.sh) | Set up webhook |

### Node.js (TypeScript)

Examples using native `fetch` (no external dependencies):

| File | Description |
|------|-------------|
| [`send-text.ts`](examples/nodejs/send-text.ts) | Send a text message |
| [`send-image.ts`](examples/nodejs/send-image.ts) | Send an image |
| [`webhook-listener.ts`](examples/nodejs/webhook-listener.ts) | Webhook server |

```bash
cd examples/nodejs
npm install
npx tsx send-text.ts
```

### Python

Examples using `requests`:

| File | Description |
|------|-------------|
| [`send_text.py`](examples/python/send_text.py) | Send a text message |
| [`send_image.py`](examples/python/send_image.py) | Send an image |
| [`webhook_listener.py`](examples/python/webhook_listener.py) | Webhook server (Flask) |

```bash
cd examples/python
pip install -r requirements.txt
python send_text.py
```

---

## Changelog

All changes and new versions are documented in [CHANGELOG.md](CHANGELOG.md).

Current version: `v3.10.0`

---

## Support

Found an issue or have a suggestion? Open an issue:

- [Report Bug](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=bug_report.yml)
- [Request Feature](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=feature_request.yml)
- [Ask a Question](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=question.yml)

For priority support, reach out via dashboard or WhatsApp.

---

## Useful Links

- [Official Website](https://zedaapi.com)
- [Interactive Documentation](https://api.zedaapi.com/docs)
- [Management Dashboard](https://zedaapi.com)
- [Platform Status](https://status.zedaapi.com)
- [n8n Node](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)

---

<p align="center">
  Made with care by <a href="https://github.com/Setup-Automatizado">Setup Automatizado</a>
</p>
