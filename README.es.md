# Zé da API - Hub de Soporte

[![Versión](https://img.shields.io/badge/version-v3.10.0-blue.svg)](https://github.com/Setup-Automatizado/use-zedaapi/releases)
[![Estado](https://img.shields.io/badge/status-activo-brightgreen.svg)](https://status.zedaapi.com)
[![Licencia](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539.svg)](openapi/openapi-latest.json)
[![n8n](https://img.shields.io/badge/n8n-community%20node-FF6D5A.svg)](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)

**[English](README.en.md)** | **Español** | **[Português](README.md)**

Hub oficial de soporte para suscriptores de **Zé da API** — la API de WhatsApp más completa de Brasil.

Aquí encontrarás la especificación OpenAPI, colección Postman, nodo n8n, ejemplos de código y el changelog de todas las versiones.

---

## Enlaces Rápidos

| Recurso | Enlace |
|---------|--------|
| Documentación API | [api.zedaapi.com/docs](https://api.zedaapi.com/docs) |
| Dashboard | [zedaapi.com](https://zedaapi.com) |
| OpenAPI JSON | [openapi-latest.json](openapi/openapi-latest.json) |
| OpenAPI YAML | [openapi-latest.yaml](openapi/openapi-latest.yaml) |
| Postman Collection | [zedaapi-latest.postman.json](postman/zedaapi-latest.postman.json) |
| n8n Community Node | [n8n-nodes-zedaapi](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Estado | [status.zedaapi.com](https://status.zedaapi.com) |

---

## Inicio Rápido

Envía tu primer mensaje en 30 segundos:

```bash
# Reemplaza con tus datos reales
HOST="https://tu-instancia.zedaapi.com"
INSTANCE_ID="tu-instancia"
INSTANCE_TOKEN="tu-token-aqui"

curl -X POST "${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}/send-text" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "message": "¡Hola! Mensaje enviado vía Zé da API."
  }'
```

> La autenticación se hace mediante **Instance ID** e **Instance Token** directamente en la URL. Consulta el dashboard para obtener tus datos.

---

## Postman Collection

Importa la colección completa en Postman para probar todas las rutas de la API:

1. Descarga el archivo [`zedaapi-latest.postman.json`](postman/zedaapi-latest.postman.json)
2. En Postman, haz clic en **Import** y selecciona el archivo
3. Configura las variables de la colección:
   - `baseUrl`: URL del servidor (ej: `https://tu-instancia.zedaapi.com`)
   - `instanceId`: ID de tu instancia
   - `instanceToken`: Token de tu instancia

La colección incluye todas las rutas organizadas por categoría: System, Instance, Messages, Webhooks, Contacts, Groups, Media y más.

---

## OpenAPI Spec

La especificación OpenAPI 3.1 está disponible en dos formatos:

- **JSON**: [`openapi/openapi-latest.json`](openapi/openapi-latest.json)
- **YAML**: [`openapi/openapi-latest.yaml`](openapi/openapi-latest.yaml)

Puedes importarla en cualquier herramienta compatible con OpenAPI (Swagger UI, Insomnia, Stoplight, etc).

Las versiones anteriores están disponibles como `openapi-vX.Y.Z.json`.

---

## n8n Community Node

Integra Zé da API directamente en tus workflows de n8n con el nodo oficial:

**[n8n-nodes-zedaapi](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)** — Envía mensajes, gestiona grupos, comunidades, newsletters, configura webhooks y más, todo desde n8n.

### Instalación en n8n

1. Ve a **Settings > Community Nodes**
2. Haz clic en **Install a community node**
3. Escribe `n8n-nodes-zedaapi` e instala

El nodo soporta todas las operaciones de la API: envío de mensajes (texto, imagen, documento, audio, video), gestión de instancias, webhooks, grupos, comunidades y newsletters.

---

## Ejemplos de Código

### cURL

Scripts shell listos para usar:

| Script | Descripción |
|--------|-------------|
| [`send-text.sh`](examples/curl/send-text.sh) | Enviar mensaje de texto |
| [`send-image.sh`](examples/curl/send-image.sh) | Enviar imagen |
| [`send-document.sh`](examples/curl/send-document.sh) | Enviar documento |
| [`instance-connect.sh`](examples/curl/instance-connect.sh) | Conectar instancia (QR Code) |
| [`webhook-setup.sh`](examples/curl/webhook-setup.sh) | Configurar webhook |

### Node.js (TypeScript)

Ejemplos usando `fetch` nativo (sin dependencias externas):

| Archivo | Descripción |
|---------|-------------|
| [`send-text.ts`](examples/nodejs/send-text.ts) | Enviar mensaje de texto |
| [`send-image.ts`](examples/nodejs/send-image.ts) | Enviar imagen |
| [`webhook-listener.ts`](examples/nodejs/webhook-listener.ts) | Servidor de webhook |

```bash
cd examples/nodejs
npm install
npx tsx send-text.ts
```

### Python

Ejemplos usando `requests`:

| Archivo | Descripción |
|---------|-------------|
| [`send_text.py`](examples/python/send_text.py) | Enviar mensaje de texto |
| [`send_image.py`](examples/python/send_image.py) | Enviar imagen |
| [`webhook_listener.py`](examples/python/webhook_listener.py) | Servidor de webhook (Flask) |

```bash
cd examples/python
pip install -r requirements.txt
python send_text.py
```

---

## Changelog

Todos los cambios y nuevas versiones están documentados en [CHANGELOG.md](CHANGELOG.md).

Versión actual: `v3.10.0`

---

## Soporte

¿Encontraste un problema o tienes una sugerencia? Abre un issue:

- [Reportar Bug](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=bug_report.yml)
- [Solicitar Feature](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=feature_request.yml)
- [Hacer una Pregunta](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=question.yml)

Para soporte prioritario, contáctanos por el dashboard o WhatsApp.

---

## Enlaces Útiles

- [Sitio Oficial](https://zedaapi.com)
- [Documentación Interactiva](https://api.zedaapi.com/docs)
- [Dashboard de Gestión](https://zedaapi.com)
- [Estado de la Plataforma](https://status.zedaapi.com)
- [Nodo n8n](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)

---

<p align="center">
  Hecho con dedicación por <a href="https://github.com/Setup-Automatizado">Setup Automatizado</a>
</p>
