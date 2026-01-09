GROUPS_zapi.md
Buscar grupos
Método#
/groups#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/groups

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por retornar todos os grupos.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
page	integer	Utilizado para paginação você de informar aqui a pagina de grupos que quer buscar
pageSize	integer	Especifica o tamanho do retorno de grupos por pagina
Opcionais#
Atributos	Tipo	Descrição
Request Params#
URL exemplo#
Método

GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/groups?page=1&pageSize=10

Response#
200#
Atributos	Tipo	Descrição
archived	boolean	true ou false indica se o chat está arquivado
pinned	boolean	true ou false indica se o chat está fixado
phone	string	Phone do contato
unread	string	indica o numero de mensagens não lidas em um chat
name	string	Nome atribudo ao chat, lembrando que se for um grupo ou lista de transmissão deve retorar os respectivos IDs
lastMessageTime	string	Timestamp com a data e hora da última interação com o chat
muteEndTime	string	Timestamp com a data e hora que a notificação vai ser reativada (-1 é para sempre)
isMuted	string	0 ou 1 indica se você silênciou ou não este chat
isMarkedSpam	boolean	true ou false indica se você marcou este chat como spam
isGroup	boolean	true ou false indica se é um grupo ou não
messagesUnread	integer	descontinuado
Exemplo

[
  {
    "isGroup": true,
    "name": "Grupo teste",
    "phone": "120263358412332916-group",
    "unread": "0",
    "lastMessageTime": "1730918668000",
    "isMuted": "0",
    "isMarkedSpam": "false",
    "archived": "false",
    "pinned": "false",
    "muteEndTime": null,
    "messagesUnread": "0"
  }
]
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Criando grupos
Método#
/create-group#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/create-group

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por criar um grupo com seus respectivos participantes. Infelizmente não é possivel criar o grupo com imagem, mas você pode logo após a criação utilizar-se do método Update-group-photo que esta nesta mesma sessão.

Dica
Assim como no WhatsApp Web você vai precisar adicionar ao menos um contato para conseguir criar um grupo.

warning
Você não deve passar o número conectado ao FUNNELCHAT que é responsável pela criação do grupo no array de números que vão compor o grupo.

Novo atributo
Recentemente, o WhatsApp implementou uma validação para verificar se o número de telefone conectado à API possui o contato do cliente salvo. No entanto, a FUNNELCHAT desenvolveu uma solução para contornar essa validação, introduzindo um novo recurso chamado "autoInvite". Agora, quando uma solicitação é enviada para adicionar 10 clientes a um grupo e apenas 5 deles são adicionados com sucesso, a API envia convites privados para os cinco clientes que não foram adicionados. Esses convites permitem que eles entrem no grupo, mesmo que seus números de telefone não estejam salvos como contatos.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
autoInvite	boolean	true ou false (Envia link de convite do grupo no privado)
groupName	string	Nome do grupo a ser criado
phones	array string	Array com os números a serem adicionados no grupo
Opcionais#
Atributos	Tipo	Descrição
Request Body#
Método

POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/create-group

Exemplo

{
  "autoInvite": true,
  "groupName": "Grupo FUNNELCHAT",
  "phones": ["5544999999999", "5544888888888"]
}
Response#
200#
Atributos	Tipo	Descrição
phone	string	ID/Fone do grupo
invitationLink	string	link para entrar no grupo
Exemplo

  {
    "phone": "120363019502650977-group",
    "invitationLink": "https://chat.whatsapp.com/GONwbGGDkLe8BifUWwLgct"
  }

405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Atualizar nome do grupo
Método#
/update-group-name#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-group-name

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável alterar o nome de um grupo já existente.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
groupName	string	Nome do grupo a ser criado
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-group-name

Body#

  {
    "groupId": "120363019502650977-group",
    "groupName": "Mudou o nome Meu grupo no FUNNELCHAT"
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Atualizar imagem do grupo
Método#
/update-group-photo#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-group-photo

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável alterar a imagem de um grupo já existente.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
groupPhoto	string	Url ou Base64 da imagem
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
Método

POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-group-photo

Body#
{
  "groupId": "string",
  "groupPhoto": "https://www.funnelchat/wp-content/themes/funnelchat/dist/images/logo.svg"
}
Enviar imagem Base64
Se você tem duvidas em como enviar uma imagem Base64 acesse o tópico mensagens "Enviar Imagem", lá você vai encontrar tudo que precisa saber sobre envio neste formato.

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Adicionar Participantes
Método#
/add-participant#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-participant

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por adicionar novos participantes ao grupo.

Novo atributo
Recentemente, o WhatsApp implementou uma validação para verificar se o número de telefone conectado à API possui o contato do cliente salvo. No entanto, a FUNNELCHAT desenvolveu uma solução para contornar essa validação, introduzindo um novo recurso chamado "autoInvite". Agora, quando uma solicitação é enviada para adicionar 10 clientes a um grupo e apenas 5 deles são adicionados com sucesso, a API envia convites privados para os cinco clientes que não foram adicionados. Esses convites permitem que eles entrem no grupo, mesmo que seus números de telefone não estejam salvos como contatos.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
autoInvite	boolean	Envia link de convite do grupo no privado
groupId	string	ID/Fone do grupo
phones	array string	Array com os número(s) do(s) participante(s) a serem adicionados
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-participant

Body#

  {
  "autoInvite": true,
  "groupId": "120363019502650977-group",
  "phones": ["5544999999999", "5544888888888"]
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Remover Participantes
Método#
/remove-participant#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-participant

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por remover participantes do grupo.
Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
phones	array string	Array com os número(s) do(s) participante(s) a serem removidos
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-participant

Body#

  {
    "groupId": "120363019502650977-group",
    "phones": ["5544999999999", "5544888888888"]
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Aprovar Participantes
Método#
/approve-participant#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/approve-participant

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por aceitar a entrada de participantes no grupo.
Atributos#
Obrigatórios#
Attributes	Type	Description
groupId	string	ID/Fone do grupo
phones	array string	Array com os número(s) do(s) participante(s) a serem aceitos
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/approve-participant

Body#

  {
    "groupId": "120363019502650977-group",
    "phones": ["5544999999999", "5544888888888"]
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Rejeitar Participantes
Método#
/reject-participant#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/reject-participant

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por rejeitar a entrada de participantes no grupo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
phones	array string	Array com os número(s) do(s) participante(s) a serem rejeitados
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/reject-participant

Body#
  {
    "groupId": "120363019502650977-group",
    "phones": ["5544999999999", "5544888888888"]
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Promover admin do grupo
Método#
/add-admin#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-admin

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por promover participamentes do grupo à administradores, você pode provomover um ou mais participamente à administrador.
Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
phones	array string	Array com os número(s) do(s) participante(s) a serem promovidos
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-admin

Body#

  {
    "groupId": "120363019502650977-group",
    "phones": ["5544999999999", "5544888888888"]
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Remover admin do grupo
Método#
/remove-admin#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-admin

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável remover um ou mais admistradores de um grupo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	id/fone do grupo
phones	array string	Array com os número(s) a ser(em) removido(s) da administração do grupo
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-admin

Body#

  {
    "groupId": "120363019502650977-group",
    "phones": ["5544999999999", "5544888888888"]
  }


---

## Response

### 200

**Exemplo**

```json
{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Sair do grupo
Método#
/leave-group#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/leave-group

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite você sair de um grupo ao qual participa.
Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/leave-group

Body#

  {
    "groupId": "120363019502650977-group"
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Metadata do Grupo
Método#
/group-metadata#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/group-metadata/{phone}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna o metadata do grupo com todas informações do grupo e de seus partipantes.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
Opcionais#
Atributos	Tipo	Descrição
Request Params#
URL#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/group-metadata/{phone}

Response#
200#
Atributos	Tipo	Descrição
phone	string	ID/Fone do Grupo
description	string	Descrição do grupo
owner	string	Número do criador do grupo
subject	string	Nome do grupo
creation	timestamp	Timestamp da data de criação do grupo
invitationLink	url	Link de convite do grupo (retorna apenas para admin)
invitationLinkError	string	Mensagem de erro caso invitationLink seja nulo (forbidden, not-authorized, too many requests)
communityId	string	ID da comunidade
adminOnlyMessage	boolean	Indica se apenas Admin pode mandar mensagens
adminOnlySettings	boolean	Indica se apenas Admin pode mudar as configurações
requireAdminApproval	boolean	Indica se necessita aprovação de admin para entrar no grupo
isGroupAnnouncement	boolean	Indica se é um grupo de aviso
participants	array string	com dados dos participantes
Array String (participants)

Atributos	Tipo	Descrição
phone	string	Fone do participante
isAdmin	string	Indica se o participante é administrador do grupo
isSuperAdmin	string	Indica se é o criador do grupo
Exemplo

  {
  "phone": "120363019502650977-group",
  "description": "Grupo FUNNELCHAT",
  "owner": "5511999999999",
  "subject": "Meu grupo no FUNNELCHAT",
  "creation": 1588721491000,
  "invitationLink": "https://chat.whatsapp.com/40Aasd6af1",
  "invitationLinkError": null,
  "communityId": null,
  "adminOnlyMessage": false,
  "adminOnlySettings": false,
  "requireAdminApproval": false,
  "isGroupAnnouncement": false,
  "participants": [
    {
      "phone": "5511888888888",
      "lid": "",
      "isAdmin": false,
      "isSuperAdmin": false
    },
    {
      "phone": "5511777777777",
      "lid": "",
      "isAdmin": true,
      "isSuperAdmin": false,
      "short": "ZAPIs",
      "name": "ZAPIs Boys"
    }
  ],
  "subjectTime": 1617805323000,
  "subjectOwner": "554497050785"
}

405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Metadata do Grupo (leve)
Método#
/light-group-metadata#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/light-group-metadata/{phone}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna o metadata do grupo com todas informações do grupo e de seus participantes com exceção do link de convite do grupo.

A única diferença entre este método e o Metadata do Grupo é que nesse não é retornado o link de convite do grupo, pois em certos momento obter esse link é custoso e ligeiramente demorado. Sabendo disso, disponibilizamos uma forma "leve" de obter o metadata do grupo.

Caso você queira utilizar este método e posteriormente necessite do link de convite do grupo, você pode obtê-lo na API de Obter link de convite do grupo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
Opcionais#
Atributos	Tipo	Descrição
Request Params#
URL#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/light-group-metadata/{phone}

Response#
200#
Atributos	Tipo	Descrição
phone	string	ID/Fone do Grupo
description	string	Descrição do grupo
owner	string	Número do criador do grupo
subject	string	Nome do grupo
creation	timestamp	Timestamp da data de criação do grupo
communityId	string	ID da comunidade
adminOnlyMessage	boolean	Indica se apenas Admin pode mandar mensagens
adminOnlySettings	boolean	Indica se apenas Admin pode mudar as configurações
requireAdminApproval	boolean	Indica se necessita aprovação de admin para entrar no grupo
isGroupAnnouncement	boolean	Indica se é um grupo de aviso
participants	array string	com dados dos participantes
Array String (participants)

Atributos	Tipo	Descrição
phone	string	Fone do participante
isAdmin	string	Indica se o participante é administrador do grupo
isSuperAdmin	string	Indica se é o criador do grupo
Exemplo

  {
  "phone": "120363019502650977-group",
  "description": "Grupo FUNNELCHAT",
  "owner": "5511999999999",
  "subject": "Meu grupo no FUNNELCHAT",
  "creation": 1588721491000,
  "invitationLink": null,
  "invitationLinkError": null,
  "communityId": null,
  "adminOnlyMessage": false,
  "adminOnlySettings": false,
  "requireAdminApproval": false,
  "isGroupAnnouncement": false,
  "participants": [
    {
      "phone": "5511888888888",
      "lid": "",
      "isAdmin": false,
      "isSuperAdmin": false
    },
    {
      "phone": "5511777777777",
      "lid": "",
      "isAdmin": true,
      "isSuperAdmin": false,
      "short": "ZAPIs",
      "name": "ZAPIs Boys"
    }
  ],
  "subjectTime": 1617805323000,
  "subjectOwner": "554497050785"
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"


Metadata do Grupo por Convite
Método#
/group-invitation-metadata#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/group-invitation-metadata?url=url-do-grupo

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna o metadata do grupo com todas informações do grupo e de seus partipantes.

Response#
200#
Atributos	Tipo	Descrição
phone	string	ID/Fone do Grupo
owner	string	Número do criador do grupo
subject	string	Nome do grupo
description	string	Descrição do grupo
creation	timestamp	Timestamp da data de criação do grupo
invitationLink	url	Link de convite do grupo
contactsCount	number	Número de contatos presente no grupo
participantsCount	number	Número de participantes no grupo
participants	array string	com dados dos participantes
Array String (participants)

Atributos	Tipo	Descrição
phone	string	Número do participante
isAdmin	string	Indica se o participante é administrador do grupo
isSuperAdmin	string	Indica se é o criador do grupo
subjectTime	timestamp	Data de criação do grupo
subjectOwner	string	Número do criador do grupo
Exemplo


  {
    "phone": "120363019502650977-group",
    "owner": "5511888888888",
    "subject": "Meu grupo no FUNNELCHAT",
    "description": "descrição do grupo",
    "creation": 1588721491000,
    "invitationLink": "https://chat.whatsapp.com/KNmcL17DqVA0sqkQ5LLA5",
    "contactsCount": 1,
    "participantsCount": 1,
    "participants": [
      {
        "phone": "5511888888888",
        "lid": "",
        "isAdmin": false,
        "isSuperAdmin": true
      },
      {
        "phone": "5511777777777",
        "lid": "",
        "isAdmin": false,
        "isSuperAdmin": false
      }
    ],
    "subjectTime": 1617805323000,
    "subjectOwner": "5511888888888"
}

405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Configurações do grupo
Método#
/update-group-settings#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-group-settings

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite você alterar as preferências do grupo.

Atenção
Atenção somente administradores podem alterar as preferências do grupo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
phone	string	ID/Fone do grupo
adminOnlyMessage	boolean	Somente administradores podem enviar mensagens no grupo
adminOnlySettings	boolean	Atributo para permitir que apenas os admins façam edições no grupo
requireAdminApproval	boolean	Define se vai ser necessário a aprovação de algum admin para entrar no grupo
adminOnlyAddMember	boolean	Somente administradores podem adicionar pessoas no grupo
Request Body#

  {
    "phone": "120363019502650977-group",
    "adminOnlyMessage": true,
    "adminOnlySettings": true,
    "requireAdminApproval": false,
    "adminOnlyAddMember": true
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Alterar descrição
Método#
/update-group-description#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-group-description

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite você alterar a descrição do grupo.

Atenção
Atenção somente administradores podem alterar as preferências do grupo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
groupDescription	string	Atributo para alterar a descrição do grupo
Body#

  {
    "groupId": "120363019502650977-group",
    "groupDescription": "descrição do grupo"
  }

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Redefinir link de convite do grupo
Método#
/redefine-invitation-link/{groupId}#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/redefine-invitation-link/{groupId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite que você redefina o link de convite de um grupo.
Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
Request url#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/redefine-invitation-link/120363019502650977-group

Response#
200#
Atributos	Tipo	Descrição
invitationLink	string	Novo link de convite
Exemplo

{
  "invitationLink": "https://chat.whatsapp.com/C1adgkdEGki7554BWDdMkd"
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Obter link de convite do grupo
Método#
/group-invitation-link/{groupId}#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/group-invitation-link/{groupId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite que você obtenha o link de convite de um grupo.
Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
groupId	string	ID/Fone do grupo
Request url#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/group-invitation-link/120363019502650977-group

Response#
200#
Atributos	Tipo	Descrição
invitationLink	string	Novo link de convite
Exemplo

{
  "phone": "120363019502650977-group",
  "invitationLink": "https://chat.whatsapp.com/C1adgkdEGki7554BWDdMkd"
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Aceitar convite do grupo
Método#
/accept-invite-group?url={{URL_DE_CONVITE}}#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/accept-invite-group?url={{URL_DE_CONVITE}}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por aceitar um convite que você recebeu para participar de um grupo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
url	string	Url recebida de convite do grupo. Pode ser obtida nesse webhook
URL#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/accept-invite-group?url=https://chat.whatsapp.com/bh8XyNrIUj84YZoy5xcaa112

Response#
200#
Atributos	Tipo	Descrição
success	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "success": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
