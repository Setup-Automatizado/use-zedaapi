# FUNNELCHAT Reference Tables

Consolida as rotas de grupos, comunidades e canais com campos obrigatórios/opcionais de request e principais atributos de resposta segundo a documentação FUNNELCHAT (`*_zapi.md`).

## Groups

### Requests

| Endpoint | Método | Payload | Campos obrigatórios | Campos opcionais |
| --- | --- | --- | --- | --- |
| `/groups` | GET | Query | `page (integer)`<br>`pageSize (integer)` | — |
| `/create-group` | POST | Body (JSON) | `autoInvite (boolean)`<br>`groupName (string)`<br>`phones ([]string)` | — |
| `/update-group-name` | POST | Body (JSON) | `groupId (string)`<br>`groupName (string)` | — |
| `/update-group-photo` | POST | Body (JSON) | `groupId (string)`<br>`groupPhoto (string URL/Base64)` | — |
| `/add-participant` | POST | Body (JSON) | `groupId (string)`<br>`phones ([]string)`<br>`autoInvite (boolean)` | — |
| `/remove-participant` | POST | Body (JSON) | `groupId (string)`<br>`phones ([]string)` | — |
| `/approve-participant` | POST | Body (JSON) | `groupId (string)`<br>`phones ([]string)` | — |
| `/reject-participant` | POST | Body (JSON) | `groupId (string)`<br>`phones ([]string)` | — |
| `/add-admin` | POST | Body (JSON) | `groupId (string)`<br>`phones ([]string)` | — |
| `/remove-admin` | POST | Body (JSON) | `groupId (string)`<br>`phones ([]string)` | — |
| `/leave-group` | POST | Body (JSON) | `groupId (string)` | — |
| `/group-metadata/{groupId}` | GET | Path | `groupId (string)` | — |
| `/light-group-metadata/{groupId}` | GET | Path | `groupId (string)` | — |
| `/group-invitation-metadata` | GET | Query | `url (string)` | — |
| `/group-invitation-link/{groupId}` | POST | Path | `groupId (string)` | — |
| `/redefine-invitation-link/{groupId}` | POST | Path | `groupId (string)` | — |
| `/update-group-settings` | POST | Body (JSON) | `phone (string)`<br>`adminOnlyMessage (boolean)`<br>`adminOnlySettings (boolean)`<br>`requireAdminApproval (boolean)`<br>`adminOnlyAddMember (boolean)` | — |
| `/update-group-description` | POST | Body (JSON) | `groupId (string)`<br>`groupDescription (string)` | — |
| `/accept-invite-group` | GET | Query | `url (string)` | — |

### Responses (200)

| Endpoint | Campos principais |
| --- | --- |
| `/groups` | `archived (boolean)`<br>`pinned (boolean)`<br>`phone (string)`<br>`unread (string)`<br>`name (string)`<br>`lastMessageTime (string timestamp)`<br>`muteEndTime (string|null)`<br>`isMuted (string)`<br>`isMarkedSpam (boolean)`<br>`isGroup (boolean)`<br>`messagesUnread (integer, depreciado)` |
| `/create-group` | `phone (string)`<br>`invitationLink (string)` |
| Mutating POSTs (`/update-group-name`, `/update-group-photo`, `/add-participant`, `/remove-participant`, `/approve-participant`, `/reject-participant`, `/add-admin`, `/remove-admin`, `/leave-group`, `/update-group-settings`, `/update-group-description`) | `value (boolean)` |
| `/group-metadata/{groupId}` | `phone (string)`<br>`description (string)`<br>`owner (string)`<br>`subject (string)`<br>`creation (timestamp)`<br>`invitationLink (string|null)`<br>`invitationLinkError (string|null)`<br>`communityId (string|null)`<br>`adminOnlyMessage (boolean)`<br>`adminOnlySettings (boolean)`<br>`requireAdminApproval (boolean)`<br>`isGroupAnnouncement (boolean)`<br>`participants (array)` com itens `phone (string)`, `isAdmin (boolean|string)`, `isSuperAdmin (boolean|string)` e campos opcionais `lid`, `short`, `name`; além de `subjectTime (timestamp)` e `subjectOwner (string)` |
| `/light-group-metadata/{groupId}` | Mesmo conjunto de campos do metadata completo, sem `invitationLink`/`invitationLinkError` |
| `/group-invitation-metadata` | `phone (string)`<br>`owner (string)`<br>`subject (string)`<br>`description (string)`<br>`creation (timestamp)`<br>`invitationLink (string)`<br>`contactsCount (number)`<br>`participantsCount (number)`<br>`participants (array)` com `phone`, `isAdmin`, `isSuperAdmin`, além de `subjectTime`, `subjectOwner` |
| `/group-invitation-link/{groupId}` | `phone (string)`<br>`invitationLink (string)` |
| `/redefine-invitation-link/{groupId}` | `invitationLink (string)` |
| `/accept-invite-group` | `success (boolean)` |

## Communities

### Requests

| Endpoint | Método | Payload | Campos obrigatórios | Campos opcionais |
| --- | --- | --- | --- | --- |
| `/communities` | POST | Body (JSON) | `name (string)` | `description (string)` |
| `/communities` | GET | Query | — | `page (integer)`<br>`pageSize (integer)` |
| `/communities/link` | POST | Body (JSON) | `communityId (string)`<br>`groupsPhones ([]string)` | — |
| `/communities/unlink` | POST | Body (JSON) | `communityId (string)`<br>`groupsPhones ([]string)` | — |
| `/communities-metadata/{communityId}` | GET | Path | `communityId (string)` | — |
| `/redefine-invitation-link/{communityId}` | POST | Path | `communityId (string)` | — |
| `/add-participant` | POST | Body (JSON) | `communityId (string)`<br>`phones ([]string)`<br>`autoInvite (boolean)` | — |
| `/remove-participant` | POST | Body (JSON) | `communityId (string)`<br>`phones ([]string)` | — |
| `/add-admin` | POST | Body (JSON) | `communityId (string)`<br>`phones ([]string)` | — |
| `/remove-admin` | POST | Body (JSON) | `communityId (string)`<br>`phones ([]string)` | — |
| `/communities/settings` | POST | Body (JSON) | `communityId (string)`<br>`whoCanAddNewGroups ("admins"|"all")` | — |
| `/communities/{communityId}` | DELETE | Path | `communityId (string)` | — |
| `/update-community-description` | POST | Body (JSON) | `communityId (string)`<br>`communityDescription (string)` | — |

### Responses (200/201)

| Endpoint | Campos principais |
| --- | --- |
| `/communities` (POST) | `id (string)`<br>`subGroups (array)` com itens `name (string)`, `phone (string)`, `isGroupAnnouncement (boolean)` |
| `/communities` (GET) | Array de objetos `name (string)`, `id (string)` |
| `/communities/link`, `/communities/unlink`, `/communities/settings` | `success (boolean)` |
| `/communities-metadata/{communityId}` | `name (string)`<br>`id (string)`<br>`description (string)`<br>`subGroups (array)` com `name`, `phone`, `isGroupAnnouncement` |
| `/redefine-invitation-link/{communityId}` | `invitationLink (string)` |
| Participant/Admin mutations (`/add-participant`, `/remove-participant`, `/add-admin`, `/remove-admin`, `/update-community-description`) | `value (boolean)` |
| `/communities/{communityId}` (DELETE) | Corpo vazio (status 200 indica sucesso) |

## Newsletters

### Requests

| Endpoint | Método | Payload | Campos obrigatórios | Campos opcionais |
| --- | --- | --- | --- | --- |
| `/create-newsletter` | POST | Body (JSON) | `name (string)` | `description (string)` |
| `/update-newsletter-picture` | POST | Body (JSON) | `id (string@newsletter)`<br>`pictureUrl (string URL/Base64)` | — |
| `/update-newsletter-name` | POST | Body (JSON) | `id (string@newsletter)`<br>`name (string)` | — |
| `/update-newsletter-description` | POST | Body (JSON) | `id (string@newsletter)`<br>`description (string)` | — |
| `/follow-newsletter` | PUT | Body (JSON) | `id (string@newsletter)` | — |
| `/unfollow-newsletter` | PUT | Body (JSON) | `id (string@newsletter)` | — |
| `/mute-newsletter` | PUT | Body (JSON) | `id (string@newsletter)` | — |
| `/unmute-newsletter` | PUT | Body (JSON) | `id (string@newsletter)` | — |
| `/delete-newsletter` | DELETE | Body (JSON) | `id (string@newsletter)` | — |
| `/newsletter/metadata/{newsletterId}` | GET | Path | `newsletterId (string@newsletter)` | — |
| `/newsletter` | GET | — | — | — |
| `/search-newsletter` | POST | Body (JSON) | `limit (number)`<br>`filters (object)` com `countryCodes ([]string)` | `view (string)`<br>`searchText (string)` |
| `/newsletter/settings/{newsletterId}` | POST | Path + Body | `newsletterId (string@newsletter)` no path<br>`reactionCodes ("basic"|"all")` no corpo | — |
| `/newsletter/accept-admin-invite/{newsletterId}` | POST | Path | `newsletterId (string@newsletter)` | — |
| `/newsletter/remove-admin/{newsletterId}` | POST | Path + Body | `newsletterId (string@newsletter)`<br>`phone (string)` | — |
| `/newsletter/revoke-admin-invite/{newsletterId}` | POST | Path + Body | `newsletterId (string@newsletter)`<br>`phone (string)` | — |
| `/newsletter/transfer-ownership/{newsletterId}` | POST | Path + Body | `newsletterId (string@newsletter)`<br>`phone (string)` | `quitAdmin (boolean)` |

### Responses (200/201)

| Endpoint | Campos principais |
| --- | --- |
| `/create-newsletter` | `id (string@newsletter)` |
| `/update-newsletter-*`, `/follow-newsletter`, `/unfollow-newsletter`, `/mute-newsletter`, `/unmute-newsletter`, `/delete-newsletter`, `/newsletter/settings/{newsletterId}`, `/newsletter/remove-admin/{newsletterId}`, `/newsletter/revoke-admin-invite/{newsletterId}` | `value (boolean)` |
| `/newsletter/metadata/{newsletterId}` | `id (string)`<br>`creationTime (timestamp)`<br>`state (string)`<br>`name (string)`<br>`description (string)`<br>`subscribersCount (string)`<br>`inviteLink (string)`<br>`verification (string)`<br>`picture (string|null)`<br>`preview (string|null)`<br>`viewMetadata (object)` com `mute ("ON"|"OFF")`, `role ("OWNER"|"SUBSCRIBER")` |
| `/newsletter` | Array com mesmos campos do metadata individual para cada canal |
| `/search-newsletter` | Objeto com `cursor (string|null)` e `data (array)` contendo itens `id`, `name`, `description`, `subscribersCount`, `picture` |
| `/newsletter/accept-admin-invite/{newsletterId}` | Corpo vazio (201 indica aceitação) |
| `/newsletter/transfer-ownership/{newsletterId}` | `value (boolean)`<br>`message (string opcional)` |
