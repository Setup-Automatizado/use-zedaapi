# Zé da API - Hub de Apoio

[![Versão](https://img.shields.io/badge/version-v3.10.0-blue.svg)](https://github.com/Setup-Automatizado/use-zedaapi/releases)
[![Status](https://img.shields.io/badge/status-ativo-brightgreen.svg)](https://status.zedaapi.com)
[![Licença](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539.svg)](openapi/openapi-latest.json)
[![n8n](https://img.shields.io/badge/n8n-community%20node-FF6D5A.svg)](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)

**[English](README.en.md)** | **[Español](README.es.md)** | **Português**

Hub oficial de apoio para assinantes do **Zé da API** — a API de WhatsApp mais completa do Brasil.

Aqui você encontra a especificação OpenAPI, coleção Postman, node n8n, exemplos de código e changelog de todas as versões.

---

## Links Rápidos

| Recurso | Link |
|---------|------|
| Documentação API | [api.zedaapi.com/docs](https://api.zedaapi.com/docs) |
| Dashboard | [zedaapi.com](https://zedaapi.com) |
| OpenAPI JSON | [openapi-latest.json](openapi/openapi-latest.json) |
| OpenAPI YAML | [openapi-latest.yaml](openapi/openapi-latest.yaml) |
| Postman Collection | [zedaapi-latest.postman.json](postman/zedaapi-latest.postman.json) |
| n8n Community Node | [n8n-nodes-zedaapi](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Status | [status.zedaapi.com](https://status.zedaapi.com) |

---

## Início Rápido

Envie sua primeira mensagem em 30 segundos:

```bash
# Substitua pelos seus dados reais
HOST="https://sua-instancia.zedaapi.com"
INSTANCE_ID="sua-instancia"
INSTANCE_TOKEN="seu-token-aqui"

curl -X POST "${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}/send-text" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "message": "Olá! Mensagem enviada via Zé da API."
  }'
```

> A autenticação é feita via **Instance ID** e **Instance Token** diretamente na URL. Consulte o dashboard para obter seus dados.

---

## Postman Collection

Importe a coleção completa no Postman para testar todas as rotas da API:

1. Baixe o arquivo [`zedaapi-latest.postman.json`](postman/zedaapi-latest.postman.json)
2. No Postman, clique em **Import** e selecione o arquivo
3. Configure as variáveis da coleção:
   - `baseUrl`: URL do servidor (ex: `https://sua-instancia.zedaapi.com`)
   - `instanceId`: ID da sua instância
   - `instanceToken`: Token da sua instância

A coleção inclui todas as rotas organizadas por categoria: System, Instance, Messages, Webhooks, Contacts, Groups, Media e mais.

---

## OpenAPI Spec

A especificação OpenAPI 3.1 está disponível em dois formatos:

- **JSON**: [`openapi/openapi-latest.json`](openapi/openapi-latest.json)
- **YAML**: [`openapi/openapi-latest.yaml`](openapi/openapi-latest.yaml)

Você pode importar em qualquer ferramenta compatível com OpenAPI (Swagger UI, Insomnia, Stoplight, etc).

Versões anteriores ficam disponíveis no formato `openapi-vX.Y.Z.json`.

---

## n8n Community Node

Integre o Zé da API diretamente nos seus workflows n8n com o node oficial:

**[n8n-nodes-zedaapi](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)** — Envie mensagens, gerencie grupos, comunidades, newsletters, configure webhooks e mais, tudo direto do n8n.

### Instalação no n8n

1. Vá em **Settings > Community Nodes**
2. Clique em **Install a community node**
3. Digite `n8n-nodes-zedaapi` e instale

O node suporta todas as operações da API: envio de mensagens (texto, imagem, documento, áudio, vídeo), gerenciamento de instâncias, webhooks, grupos, comunidades e newsletters.

---

## Exemplos de Código

### cURL

Scripts shell prontos para usar:

| Script | Descrição |
|--------|-----------|
| [`send-text.sh`](examples/curl/send-text.sh) | Enviar mensagem de texto |
| [`send-image.sh`](examples/curl/send-image.sh) | Enviar imagem |
| [`send-document.sh`](examples/curl/send-document.sh) | Enviar documento |
| [`instance-connect.sh`](examples/curl/instance-connect.sh) | Conectar instância (QR Code) |
| [`webhook-setup.sh`](examples/curl/webhook-setup.sh) | Configurar webhook |

### Node.js (TypeScript)

Exemplos usando `fetch` nativo (sem dependências externas):

| Arquivo | Descrição |
|---------|-----------|
| [`send-text.ts`](examples/nodejs/send-text.ts) | Enviar mensagem de texto |
| [`send-image.ts`](examples/nodejs/send-image.ts) | Enviar imagem |
| [`webhook-listener.ts`](examples/nodejs/webhook-listener.ts) | Servidor de webhook |

```bash
cd examples/nodejs
npm install
npx tsx send-text.ts
```

### Python

Exemplos usando `requests`:

| Arquivo | Descrição |
|---------|-----------|
| [`send_text.py`](examples/python/send_text.py) | Enviar mensagem de texto |
| [`send_image.py`](examples/python/send_image.py) | Enviar imagem |
| [`webhook_listener.py`](examples/python/webhook_listener.py) | Servidor de webhook (Flask) |

```bash
cd examples/python
pip install -r requirements.txt
python send_text.py
```

---

## Changelog

Todas as alterações e novas versões estão documentadas no [CHANGELOG.md](CHANGELOG.md).

Versão atual: `v3.10.0`

---

## Suporte

Encontrou um problema ou tem uma sugestão? Abra uma issue:

- [Reportar Bug](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=bug_report.yml)
- [Solicitar Feature](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=feature_request.yml)
- [Tirar Dúvida](https://github.com/Setup-Automatizado/use-zedaapi/issues/new?template=question.yml)

Para suporte prioritário, entre em contato pelo dashboard ou WhatsApp.

---

## Links Úteis

- [Site Oficial](https://zedaapi.com)
- [Documentação Interativa](https://api.zedaapi.com/docs)
- [Dashboard de Gerenciamento](https://zedaapi.com)
- [Status da Plataforma](https://status.zedaapi.com)
- [Node n8n](https://github.com/Setup-Automatizado/n8n-nodes-zedaapi)

---

<p align="center">
  Feito com dedicação por <a href="https://github.com/Setup-Automatizado">Setup Automatizado</a>
</p>
