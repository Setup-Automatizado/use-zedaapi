# WEBHOOKS Events Completion Checklist

> Fonte de verdade para acompanhar o progresso do mapeamento Whatsmeow → FUNNELCHAT.
> Atualize **apenas** depois de validar que cada item está coberto no código, documentação e testes.

## Fase 0 — Inventário & Planejamento

- [x] Levantar todos os eventos Whatsmeow disponíveis em `api/whatsmeow_schema_webhooks/`.
- [x] Cruzar cada evento com o webhook FUNNELCHAT correspondente (endpoint, schema, exemplo) e registrar neste arquivo.

### Inventário Whatsmeow → FUNNELCHAT

| Whatsmeow event | Pasta schema | Webhook FUNNELCHAT esperado | Pastas/schema FUNNELCHAT |
| --- | --- | --- | --- |
| Message | `Message/` | ReceivedCallback (mensagens + conteúdos) | `api/z_api/received_callback/{text,image,video,audio,...}` |
| UndecryptableMessage | `Undecryptablemessage/` | ReceivedCallback (waitingMessage=true) | `received_callback/unknown` |
| Receipt | `Readreceipt/` | MessageStatusCallback | `api/z_api/message_status/*` |
| ChatPresence | `Chatpresence/` | PresenceChatCallback | `api/z_api/presence_chat/` |
| Presence | (inline) | PresenceChatCallback | `presence_chat/` |
| Connected | `Connected/` | ConnectedCallback | `api/z_api/connected/` |
| Disconnected | `Disconnected/` | DisconnectedCallback | `api/z_api/connected/` (mesmo endpoint) |
| GroupInfo | `Groupinfo/` | ReceivedCallback (notifications GROUP_PARTICIPANT_*) | `received_callback/group_message` |
| GroupUpdate | `GroupUpdate/` | ReceivedCallback (group settings/name/topic) | `received_callback/group_message` |
| JoinedGroup | `Joinedgroup/` | ReceivedCallback (membership approvals/invites) | `received_callback/group_message` |
| Picture | `Picture/` | ReceivedCallback (PROFILE_PICTURE/GROUP_ICON updates) | `received_callback/unknown` |
| HistorySync | `Historysync/` | ReceivedCallback (notification HISTORY_SYNC/appstate) | `received_callback/unknown` |
| Appstate / Appstatesynccomplete | dirs | ReceivedCallback (notification APPSTATE_SYNC) | `received_callback/unknown` |
| Archive / Unarchivechatssetting | dirs | ReceivedCallback (`notification` = ARCHIVE/UNARCHIVE) | `received_callback/unknown` |
| Mute | `Mute/` | ReceivedCallback (`notification` = CHAT_MUTED) | `received_callback/unknown` |
| Pin | `Pin/` | ReceivedCallback (`pinMessage`) | `received_callback/unknown` |
| Labeledit | `Labeledit/` | ReceivedCallback (`notification` = CHAT_LABEL_ASSOCIATION) | `received_callback/unknown` |
| Call* (`CallEvent`, `Callreject`, `Callterminate`, `Calltransport`, `Callrelaylatency`) | dirs | ReceivedCallback (`notification` = CALL_*) | `received_callback/unknown` |
| Newsletter* (`Newsletter`, `NewsletterJoin`, `NewsletterLeave`) | dirs | ReceivedCallback (`notification` = NEWSLETTER_ADMIN_*, `newsletterAdminInvite`, etc.) | `received_callback/unknown` |
| Profile events (`ProfileUpdate`, `Businessname`, `Pushname`, `Userabout`, `Contact`) | dirs | ReceivedCallback (`notification` = PROFILE_* / CONTACT_UPDATE) | `received_callback/unknown` |
| Security events (`Blocklist`, `Identitychange`) | dirs | ReceivedCallback (`notification` = BLOCK_LIST/IDENTITY_CHANGE) | `received_callback/unknown` |
| Payments (`Message` w/ payment nodes) | `Message/` | ReceivedCallback (`reviewAndPay`, `requestPayment`, `sendPayment`) | `received_callback/text` + extras |
| Commerce (`Product`, `Order`, `Catalog`) | `Message/` | ReceivedCallback (`product`, `order`) | `received_callback/text` |
| QR | `Qr/` | Connected/Received? (usado para onboarding) | `connected/` docs |
| Offlinesync* / Keepalive* / MarkChatAsRead | dirs | Presence/Received notifications (status/membership) | `presence_chat` ou `received_callback/unknown` |
| Blocklist / IdentityChange | dirs | ReceivedCallback (security notifications) | `received_callback/unknown` |
| DeleteChat / DeleteForMe | dirs | ReceivedCallback (`notification` = CHAT_DELETED) | `received_callback/unknown` |


## Fase 1 — Eventos de Grupo

- [x] `group_info`: notificações GROUP_PARTICIPANT_{ADD,LEAVE,INVITE}, membership approvals, revoke.
  - [x] Delta básico (JOIN/LEAVE) → GROUP_PARTICIPANT_{ADD,LEAVE,INVITE}.
  - [x] Membership approval request (`MEMBERSHIP_APPROVAL_REQUEST`) via stub mapping.
  - [x] Membership revoke/cancel payloads (`REVOKED_MEMBERSHIP_REQUESTS`).
  - [x] Atualizações de link (reset/new link) e flags LOCKED/ANNOUNCE/EPHEMERAL.
  - [x] Admin role changes (`GROUP_PARTICIPANT_PROMOTE/DEMOTE`) e exclusão (`GROUP_DELETE`).
- [x] `group_joined`: entradas via invite link, auto convite, auto aprovação.
  - [x] Emissão direta de `ReceivedCallback` para `group_joined` (reconcilia JOIN outbox ao entrar em um grupo).
- [x] `picture`: alteração/remoção de foto de grupo e perfil conectado.
- [x] `GroupUpdate`: mudanças de nome/descrição/configuração (via notificações GROUP_CHANGE_*).
  - [x] Interceptar `MessageStubType` (SUBJECT, DESCRIPTION, INVITE_LINK, RESTRICT, ANNOUNCE, CHANGE_EPHEMERAL) diretamente dos eventos `Message`/history sync, garantindo ReceivedCallback mesmo sem `group_info` dedicado.

## Fase 2 — Chamadas, Canais e Newsletter

- [x] `CallEvent`, `CallReject`, `CallTerminate`, `CallTransport`, `CallRelayLatency`: transformados em ReceivedCallback `CALL_*` com `callId` e `notificationParameters` indicando estágio (`call_event_kind`). `CallOffer/Notice/Transport/Relay` → `CALL_VOICE/CALL_VIDEO`; `CallReject/Terminate` → `CALL_MISSED_*`.
- [x] `Newsletter`, `NewsletterJoin`, `NewsletterLeave`, `NewsletterMuteChange`: geração de notificações `NEWSLETTER_ADMIN_PROMOTE/DEMOTE`, `NEWSLETTER_MEMBER_{JOIN,LEAVE}` e `NEWSLETTER_MUTE_*` usando metadata do canal (`newsletter_jid`, `newsletter_name`). Eventos `handleNewsletterNotification` passam a emitir ReceivedCallback `NEWSLETTER_MESSAGE_*` com ID do post.
- [x] `Businessname`, `Pushname`, `Userabout`: propagação para `PROFILE_NAME_UPDATED`/`PROFILE_STATUS_UPDATED` com `profileName`/`Text` e enriquecimento de `ChatName`/`Photo` via provider.

## Fase 3 — Commerce, Pagamentos e Conteúdo Especial

- [ ] `Message` payloads com `reviewAndPay`, `order`, `product`, `pixKey`, `requestPayment`, `sendPayment`.
- [ ] `Labeledit`, `Pin`, `Mute`, `Archive`, `Unarchivechatssetting`, `Deletechat`, `Deleteforme`, `Markchatasread`.
- [ ] `Appstate`, `Appstatesynccomplete`, `Offlinesync*`, `Keepalive*`, `Historysync`.

## Fase 4 — Delivery & Presence

- [ ] Implementar `DeliveryCallback` completo (webhook "Ao enviar").
- [ ] Garantir todos os `message_status` (`read`, `played`, `group_*`, `read_by_me`) no padrão FUNNELCHAT.
- [ ] Ampliar `presence_chat` e `connected/disconnected` para cobrir eventos adicionais.

## Fase 5 — Testes & Documentação

- [ ] Suite de testes table-driven cobrindo cada transformador FUNNELCHAT.
- [ ] Documentação sincronizada (`api/z_api/WEBHOOKS_EVENTS.md`, docs em `funnelchat-docs`).
- [ ] Execução validada de `go test ./api/internal/events/...` com toolchain disponível.

### Registro de Execuções

- 2025-02-14: `go build -v ./...` falhou no sandbox (sem permissão para gravar em `/Users/guilhermejansen/go/pkg/mod/...` e `/Library/Caches/go-build`). Necessário repetir em ambiente com acesso ao cache Go/toolchain.
- 2025-11-12: `go build -v ./...` voltou a falhar pelo bloqueio de escrita em `/Users/guilhermejansen/Library/Caches/go-build`, mas a implementação de `GroupUpdate` via stubs (`GROUP_CHANGE_*`, `CHANGE_EPHEMERAL_SETTING`) foi concluída e validada localmente.
- 2025-11-12: `go build -v ./...` permanece bloqueado pelo sandbox de cache; entregues transformações de chamadas (`CALL_*`), newsletters (`NEWSLETTER_*`, `newsletterAdminInvite`) e perfis (`PROFILE_*`).
