# Status de Implementação dos Endpoints Z-API

**Última atualização:** 2025-01-11

## Legenda
- ✅ **Implementado e Funcional** - Endpoint completo com handler, service e OpenAPI docs
- ⚠️ **Parcialmente Implementado** - Funcionalidade existe mas precisa ajustes
- ❌ **Não Implementado** - Precisa ser criado do zero
- 🔄 **Em Progresso** - Atualmente sendo desenvolvido

---

## 📱 Instance

| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Dados do celular | GET /device | ✅ | instances.go | Resposta compatível com Z-API |
| Status da instância | GET /status | ✅ | instances.go | Completo |
| QR Code | GET /qr-code | ✅ | instances.go | Completo |
| Phone Code | GET /phone-code | ✅ | instances.go | Completo |

---

## 📨 Messages

### Envio Básico
| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Enviar texto | POST /send-text | ✅ | messages.go | Completo com OpenAPI docs |
| Enviar imagem | POST /send-image | ✅ | messages.go | Completo com OpenAPI docs |
| Enviar sticker | POST /send-sticker | ✅ | messages.go | Completo com OpenAPI docs |
| Enviar GIF | POST /send-gif | ✅ | messages.go | Completo com OpenAPI docs |
| Enviar áudio | POST /send-audio | ✅ | messages.go | Completo |
| Enviar vídeo | POST /send-video | ✅ | messages.go | Completo |
| Enviar documento | POST /send-document | ✅ | messages.go | Implementado como /send-document |
| Enviar localização | POST /send-location | ✅ | messages.go | Completo |
| Enviar contato | POST /send-contact | ✅ | messages.go | Completo |
| Enviar contatos | POST /send-contacts | ✅ | messages.go | Completo |

### Envio Avançado
| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Enviar PTV | POST /send-ptv | ❌ | - | Push-to-talk video - precisa implementar |
| Enviar link | POST /send-link | ❌ | - | Preview de link - precisa implementar |
| Enviar enquete | POST /send-poll | ❌ | - | Polls - precisa implementar |
| Enviar evento | POST /send-event | ❌ | - | Calendar events - precisa implementar |

### Interativos (⚠️ ATENÇÃO ESPECIAL)
| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Botões de ação | POST /send-button-actions | ❌ | - | **Usar exemplos_handlers/send_handlers.go** |
| Lista de botões | POST /send-button-list | ❌ | - | **Usar exemplos_handlers/send_handlers.go** |

**⚠️ IMPORTANTE - Botões e Listas:**
- ✅ Lógica de envio JÁ EXISTE em `send.go` (buttons, lists, carousel funcionam)
- ✅ Exemplos REAIS em `exemplos_handlers/send_handlers.go`
- ❌ Precisa criar handlers HTTP com formato Z-API
- 🎯 Request/Response: IDÊNTICO ao Z-API
- 🎯 Envio WhatsApp: IDÊNTICO aos exemplos (send_handlers.go)

### Operações de Mensagem
| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Reencaminhar | POST /forward-message | ❌ | - | Precisa implementar |
| Reagir | POST /send-reaction | ❌ | - | Emoji reactions - precisa implementar |
| Remover reação | POST /send-remove-reaction | ❌ | - | Precisa implementar |
| Deletar mensagem | DELETE /messages | ❌ | - | Query params: phone, messageId, owner |

---

## 👥 Contacts

| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Listar contatos | GET /contacts | ✅ | messages.go | Completo com paginação e OpenAPI docs |
| Metadata do contato | GET /contacts/{PHONE} | ❌ | - | Detalhes individuais - precisa implementar |
| Foto do perfil | GET /profile-picture | ❌ | - | Query param: phone - precisa implementar |
| Número tem WhatsApp? | GET /phone-exists/{PHONE} | ❌ | - | Validação individual - precisa implementar |
| Validação em lote | POST /phone-exists-batch | ❌ | - | Validação múltipla - precisa implementar |

---

## 💬 Chats

| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Listar chats | GET /chats | ✅ | messages.go | Completo com paginação e OpenAPI docs |
| Metadata do chat | GET /chats/{PHONE} | ❌ | - | Detalhes individuais - precisa implementar |

### Operações de Chat (POST /modify-chat)
| Operação | Status | Observações |
|----------|--------|-------------|
| Ler chat | ❌ | Marcar como lido - precisa implementar |
| Arquivar chat | ❌ | Archive/unarchive - precisa implementar |
| Fixar chat | ❌ | Pin/unpin - precisa implementar |
| Mutar chat | ❌ | Mute/unmute - precisa implementar |
| Limpar chat | ❌ | Clear messages - precisa implementar |
| Deletar chat | ❌ | Delete conversation - precisa implementar |

### Outras Operações de Chat
| Endpoint | Método | Status | Observações |
|----------|--------|--------|---------|
| Expiração de chats | POST /send-chat-expiration | ❌ | Disappearing messages - precisa implementar |

---

## 📞 Calls

| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Fazer ligação | POST /send-call | ❌ | - | Voice/Video calls - precisa implementar |

---

## 📱 Status (Stories)

| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Texto status | POST /send-text-status | ❌ | - | Text story - precisa implementar |
| Imagem status | POST /send-image-status | ❌ | - | Image story - precisa implementar |
| Áudio status | POST /send-audio-status | ❌ | - | Audio story - precisa implementar |

---

## 📊 Queue (Fila de Mensagens)

| Endpoint | Método | Status | Arquivo | Observações |
|----------|--------|--------|---------|-------------|
| Listar fila | GET /queue | ✅ | messages.go | Paginação completa |
| Contagem da fila | GET /queue/count | ✅ | messages.go | Total de mensagens |
| Limpar fila | DELETE /queue | ✅ | messages.go | Remove todas |
| Cancelar mensagem | DELETE /queue/{zaapId} | ✅ | messages.go | Remove individual |

---

## 📝 Resumo Executivo

### ✅ Implementados (22 endpoints)
- **Instance:** 3/4 (75%)
- **Messages Básico:** 10/10 (100%)
- **Contacts:** 1/5 (20%)
- **Chats:** 1/8 (12.5%)
- **Queue:** 4/4 (100%)

### ❌ Faltantes (31 endpoints)
- **Instance:** 1 endpoint (device info)
- **Messages Avançado:** 4 endpoints (PTV, link, poll, event)
- **Messages Interativos:** 2 endpoints (button-actions, button-list) - **PRIORIDADE COM ATENÇÃO ESPECIAL**
- **Messages Operações:** 3 endpoints (forward, reactions, delete)
- **Contacts:** 4 endpoints (metadata, profile-picture, validation)
- **Chats:** 7 endpoints (metadata, modify operations, expiration)
- **Calls:** 1 endpoint (send-call)
- **Status:** 3 endpoints (text, image, audio stories)

### 🎯 Prioridades Sugeridas

**Prioridade ALTA (Funcionalidade Core):**
1. ⚠️ Botões e Listas (`/send-button-actions`, `/send-button-list`)
   - Usar `exemplos_handlers/send_handlers.go` como referência
   - Manter compatibilidade Z-API request/response
   - Lógica de envio já existe em `send.go`

2. Reações (`/send-reaction`, `/send-remove-reaction`)
3. Reencaminhar mensagem (`/forward-message`)
4. Deletar mensagem (`DELETE /messages`)

**Prioridade MÉDIA (Metadata e Validação):**
5. Metadata de contato (`GET /contacts/{PHONE}`)
6. Validação de números (`GET /phone-exists`, `POST /phone-exists-batch`)
7. Metadata de chat (`GET /chats/{PHONE}`)
8. Operações básicas de chat (ler, arquivar, fixar, mutar)

**Prioridade BAIXA (Features Avançadas):**
9. PTV, Link preview, Polls, Events
10. Status/Stories
11. Calls
12. Device info

---

## 🔧 Notas Técnicas

### Arquivos de Referência
- **Handlers:** `/api/internal/http/handlers/messages.go` (22+ funções)
- **Service:** `/api/internal/messages/service.go`
- **Client Provider:** `/api/internal/messages/client_provider.go`
- **Exemplos WhatsApp:** `/api/exemplos_handlers/send_handlers.go` ⚠️

### Padrões Estabelecidos
- ✅ Clean Architecture (Handler → Service → ClientProvider)
- ✅ Paginação (page, pageSize, X-Total-Count)
- ✅ Autenticação (Client-Token header + instance token)
- ✅ OpenAPI 3.1.0 documentation
- ✅ FIFO queue per recipient
- ✅ Observabilidade completa (logs, métricas, Sentry)

### Dependências whatsmeow
- ✅ `send.go` - Envio de mensagens (buttons, lists, carousel já funcionam)
- ✅ `group.go` - Operações de grupos
- ✅ `newsletter.go` - Canais/newsletters
- ✅ `store/` - Persistência e cache
- ⚠️ Algumas features podem exigir queries Mex customizadas

---

## 📚 Documentação

### OpenAPI Docs Completos
- ✅ POST /send-text
- ✅ POST /send-image
- ✅ POST /send-sticker
- ✅ POST /send-gif
- ✅ GET /contacts
- ✅ GET /chats

### Precisa Documentar
- ❌ Todos os 31 endpoints faltantes
- ❌ Schemas para botões/listas (usar Z-API como referência)
- ❌ Schemas para polls, events, status
- ❌ Error responses específicos de cada operação

---

**Fim do relatório. Lista criada em:** 2025-01-11
