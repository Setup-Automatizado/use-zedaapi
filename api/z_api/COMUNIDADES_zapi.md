COMUNIDADES_zapi.md
Introdução
Conceituação#
O WhatsApp agora oferece a funcionalidade de Comunidades, que permite aos usuários agrupar grupos ao redor de um assunto ou interesse em comum. É uma maneira fácil de conectar com outras pessoas que compartilham dos mesmos objetivos e ideias.

Para exemplificar como funciona a estrutura de comunidades, veja a imagem abaixo:

![Estrutura de Comunidades](whatsapp-api-golang/api/z_api/communities.png)
Observações:

Ao criar uma comunidade, é criado um grupo padrão (grupo de avisos) com o mesmo nome da comunidade.
Este grupo representa sua comunidade inteira e é usado para enviar mensagens para todos.
Cada novo grupo vinculado a comunidade, todos participantes farão parte do grupo padrão (grupo de avisos).
Ao desvincular um grupo, todos participantes dele são removidos do grupo padrão (grupo de avisos).
Como pode ser visto acima, toda comunidade possui um "Grupo de avisos" neste grupo apenas administradores podem enviar mensagens, utilize ele sempre que quiser enviar algo para toda comunidade.

Cada comunidade pode ter até 50 grupos, e o(s) administrador(es) das comunidades poderão disparar mensagens para até 5 mil pessoas de uma única vez através do grupo de avisos.

Perguntas sobre funcionamento das APIs#
1. Como faço para criar uma nova comunidades?#
Primeiro é importante verificar se o aplicativo do whatsapp do seu celular já está compatível com as comunidades, caso não esteja, aguarde a atualização do aplicativo para sua conta, agora caso já tenha acesso a comunidades, veja a documentação de como criar comunidade via API.

2. Consigo listar as comunidades que faço parte?#
Sim, o FUNNELCHAT disponibiliza os métodos para que você consiga saber quais comunidades você faz parte, veja a documentação de como listar suas comunidades.

3. Consigo vincular e desvincular grupos a uma comunidade?#
Com certeza! o FUNNELCHAT te entrega outras duas APIs para que você consiga gerenciar os grupos de uma comunidade, veja como vincular grupos ou desvincular grupos de uma comunidade.

4. Como enviar mensagem para toda comunidade?#
Como dito acima, a comunidade em si serve apenas para agrupar seus grupos e dar uma experiência e visão de todos os grupos da comunidade aos usuários. Você pode sim enviar mensagem para toda comunidade, más para isso é utilizado o Grupo anúncios. Como o grupo de avisos se trata de um grupo como qualquer outro, basta você possuir o phone do grupo e utilizar as APIs de envio de mensagem normalmente, assim como outros grupos comuns.

5. Como consigo pegar os grupo de avisos?#
Existem 3 formas de se pegar os grupos de anúncio.
- A primeira é na criação da comunidade, que ao criar a comunidade já te retorna as informações do grupo de avisos.
- A segunda é pela API de listar chats, nela você pode diferenciar grupos normais de grupos de anúncios, o atributo isGroup estará como verdadeiro sempre que se tratar de um grupo normal e o atributo isGroupAnnouncement estará verdadeiro quando for um grupo de avisos.
- A terceira e última opcão é pela API de metadata da comunidade, ela te retorna informações sobre a comunidade baseado no ID dela, retornando informações como nome da comunidade e seus grupos vinculados.

6. Posso desativar uma comunidade?#
Sim, você pode desativar uma Comunidade no WhatsApp, o que resultará na desconexão de todos os grupos relacionados a ela. É importante ressaltar que desativar a Comunidade não excluirá seus grupos, mas sim os removerá da Comunidade em questão.

7. Como adicionar ou remover pessoas da comunidade?#
Como comentado anteriormente, a comunidade em si é apenas o que agrupa seus grupos, o que de fato é utilizado são os grupos de anúncios, então caso queira gerar o link de convite, adicionar e remover pessoas, promover como administradoras etc... tudo será feito através do grupo de avisos utilizando as APIs que você já conhece.

Criar comunidade
Método#
/create-group#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Antes de utilizar esse recurso, é importante verificar se o aplicativo do WhatsApp no seu celular já possui compatibilidade com as comunidades, caso já esteja disponível você pode utilizar essa API para criar novas comunidades.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
name	string	Nome da comunidade que deseja criar
Opcionais#
Atributos	Tipo	Descrição
description	string	Descrição da comunidade
Request Body#
Método

POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities

Exemplo

{
  "name": "Minha primeira Comunidade"
}
Response#
200#
Atributos	Tipo	Descrição
id	string	ID da comunidade criada
subGroups	array[subgroup]	Lista de grupos vinculados
Exemplo

{
  "id": "98372465382764532938",
  "subGroups": [
    {
      "name": "Minha primeira Comunidade",
      "phone": "342532456234453-group",
      "isGroupAnnouncement": true
    }
  ]
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Listar comunidades
Método#
/communities#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por retornar todas as comunidades que você faz parte.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
Opcionais#
Atributos	Tipo	Descrição
page	integer	Utilizado para paginação você de informar aqui a pagina de comunidades que quer buscar
pageSize	integer	Especifica o tamanho do retorno de comunidades por pagina
Request Params#
URL exemplo#
Método

GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities

Response#
200#
Atributos	Tipo	Descrição
name	string	Nome da comunidade
id	string	Identificador da comunidade
Exemplo

[
  {
    "name": "Minha primeira Comunidade",
    "id": "98372465382764532938"
  }
]
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Vincular grupos
Método#
/communities/link#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/link

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Com essa API você consegue adicionar outros grupos a uma comunidade, para isso você vai precisar do ID da sua comunidade e os telefones dos grupos que deseja adicionar.

Atenção
É importante lembrar que não é possível vincular o mesmo grupo em mais de uma comunidade, caso você informe 3 grupos para adicionar onde 1 já esteja em uma comunidade, 2 serão adicionados e o outro retornará na resposta que já faz parte de outra comunidade.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID da comunidade que será adicionado os grupos
groupsPhones	array string	Array com os número(s) dos grupos a serem vinculados
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/link

Body#
{
  "communityId": "98372465382764532938",
  "groupsPhones": ["1345353454354354-group", "1203634230225498-group"]
}
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

Desvincular grupos
Método#
/communities/unlink#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/unlink

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Com essa API você consegue remover grupos de uma comunidade, para isso você vai precisar do ID da sua comunidade e os telefones dos grupos que deseja remover.

Atenção
Uma comunidade deve ter no mínimo 1 grupo vinculado a ela, isso sem contar com o grupo de avisos, então caso sua comunidade só possua um grupo comum vinculado, não será possível remove-lo.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID da comunidade que será desvinculado os grupos
groupsPhones	array string	Array com os número(s) dos grupos a serem desvinculados
Opcionais#
Atributos	Tipo	Descrição
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/unlink

Body#
{
  "communityId": "98372465382764532938",
  "groupsPhones": ["1345353454354354-group"]
}
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

Metadata da comunidade
Método#
/communities-metadata/{communityId}#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities-metadata/{idDaComunidade}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna o metadata da comunidade, como nome, descrição e grupos que estão vinculados a ela.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
Opcionais#
Atributos	Tipo	Descrição
Request Params#
URL#
GET https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities-metadata/{idDaComunidade}

Response#
200#
Atributos	Tipo	Descrição
name	string	Nome da comunidade
id	string	ID da comunidade
description	string	Descrição da comunidade
subGroups	array subgroup	lista de grupos vinculados
Array (subGroups)

Atributos	Tipo	Descrição
name	string	Nome do sub grupo
phone	string	Telefone do sub grupo
isGroupAnnouncement	boolean	Informe se é um grupo de avisos ou comum
Exemplo

{
  "name": "Minha primeira Comunidade",
  "id": "98372465382764532938",
  "description": "Uma descrição da comunidade",
  "subGroups": [
    {
      "phone": "Minha primeira Comunidade",
      "phone": "342532456234453-group",
      "isGroupAnnouncement": true
    },
    {
      "phone": "Outro grupo",
      "phone": "1203634230225498-group",
      "isGroupAnnouncement": false
    }
  ]
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Redefinir link de convite da comunidade
Método#
/redefine-invitation-link/{communityId}#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/redefine-invitation-link/{communityId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite que você redefina o link de convite de uma comunidade.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID/Fone da comunidade
Request url#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/redefine-invitation-link/120363019502650977

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

Adicionar participantes
Método#
/add-participant#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-participant

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por adicionar novos participantes a comunidade.

Novo atributo
Recentemente, o WhatsApp implementou uma validação para verificar se o número de telefone conectado à API possui o contato do cliente salvo. No entanto, a FUNNELCHAT desenvolveu uma solução para contornar essa validação, introduzindo um novo recurso chamado "autoInvite". Agora, quando uma solicitação é enviada para adicionar 10 clientes a um grupo e apenas 5 deles são adicionados com sucesso, a API envia convites privados para os cinco clientes que não foram adicionados. Esses convites permitem que eles entrem na comunidade, mesmo que seus números de telefone não estejam salvos como contatos.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
autoInvite	boolean	Envia link de convite da comunidade no privado
communityId	string	ID/Fone da comunidade. Pode ser obtido na API de Listar comunidades
phones	array string	Array com os número(s) do(s) participante(s) a serem adicionados
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-participant

Body#
  {
  "autoInvite": true,
  "communityId": "120363019502650977",
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

Remover participantes
Método#
/remove-participant#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-participant

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é reponsável por remover participantes da comunidade.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID/Fone da comunidade. Pode ser obtido na API de Listar comunidades
phones	array string	Array com os número(s) do(s) participante(s) a serem removidos
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-participant

Body#
  {
    "communityId": "5511999999999",
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

Promover admin da comunidade
Método#
/add-admin#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-admin

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por promover participamentes da comunidade à administradores, você pode provomover um ou mais participamente à administrador.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID/Fone da comunidade. Pode ser obtido na API de Listar comunidades
phones	array string	Array com os número(s) do(s) participante(s) a serem promovidos
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/add-admin

Body#
  {
    "communityId": "120363186053925765",
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

Remover admin da comunidade
Método#
/remove-admin#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-admin

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável remover um ou mais admistradores de uma comunidade.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	id/fone da comunidade. Pode ser obtido na API de Listar comunidades
phones	array string	Array com os número(s) a ser(em) removido(s) da administração do grupo
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/remove-admin

Body#
  {
    "communityId": "120363019502650977",
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

Configurações da comunidade
Método#
/communities/settings#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/settings

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Com essa API você consegue alterar as configurações de uma comunidade.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID da comunidade que será alterado as configurações
whoCanAddNewGroups	string (admins ou all)	Configuração de quem pode adicionar novos grupos a essa comunidade. Somente administradores (admins) ou todos (all)
Request Body#
URL#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/settings

Body#
{
  "communityId": "98372465382764532938",
  "whoCanAddNewGroups": "admins" | "all"
}
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

Desativar comunidade
Método#
/queue#
DELETE https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/{idDaComunidade}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por desativar uma comunidade.

Quando uma comunidade é desativada resultará na desconexão de todos os grupos relacionados a ela. É importante ressaltar que desativar a Comunidade não excluirá seus grupos, mas sim os removerá da Comunidade em questão.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
Opcionais#
Atributos	Tipo	Descrição
Request Params#
URL exemplo#
Método

DELETE https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/communities/{idDaComunidade}

Response#
200#
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Alterar descrição
Método#
/update-community-description#
POST https://api.funnelchat/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-community-description

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método permite você alterar a descrição da comunidade.

Atenção
Atenção somente administradores podem alterar as preferências da comunidade.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
communityId	string	ID/Fone do grupo
communityDescription	string	Atributo para alterar a descrição da comunidade
Body#

  {
    "communityId": "120363019502650977",
    "communityDescription": "descrição da comunidade"
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
