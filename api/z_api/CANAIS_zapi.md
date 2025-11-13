CANAIS_zapi.md
Introdução
Conceituação#
Neste tópico você vai conhecer todos os métodos disponiveis para manipulação de canais.

O que são os canais? Os Canais do WhatsApp são semelhantes às páginas do Facebook, oferecendo um espaço para compartilhar informações, novidades, avisos e outros conteúdos relacionados a um tópico específico ou entidade. Esses canais podem ser criados por uma variedade de entidades, desde times de futebol até empresas que oferecem serviço de streaming, e permitem que os usuários acompanhem e se envolvam com o conteúdo de seu interesse.

ID de Canais
É necessario enfatizar que o padrão utilizado pelo WhatsApp para os ids dos canais é sempre com o sufixo "@newsletter".

Criando canais
Método#
/create-newsletter#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/create-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por criar um canal. Infelizmente não é possivel criar o canal com imagem, mas você pode logo após a criação utilizar-se do método update-newsletter-picture que esta nesta mesma sessão.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
name	string	Nome do canal
Opcionais#
Atributos	Tipo	Descrição
description	string	Descrição do canal
Request Body#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/create-newsletter

Exemplo

{
  "name": "Nome do canal",
  "description": "Descrição do canal"
}
Response#
201#
Atributos	Tipo	Descrição
id	string	ID do canal
Exemplo


  {
    "id": "999999999999999999@newsletter",
  }

tip
O id retornado sempre conterá o sufixo "@newsletter", padrão utilizado pelo próprio WhatsApp. Ele deve ser utilizado no mesmo formato nas requisições que recebem o id como parâmetro.

405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Atualizar imagem do canal
Método#
/update-newsletter-picture#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-newsletter-picture

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por alterar a imagem de um canal já existente.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
pictureUrl	string	Url ou Base64 da imagem
Request Body#
URL#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-newsletter-picture

Body#
{
  "id": "999999999999999999@newsletter",
  "pictureUrl": "https://www.z-api.io/wp-content/themes/z-api/dist/images/logo.svg"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Enviar imagem Base64
Se você tem duvidas em como enviar uma imagem Base64 acesse o tópico mensagens "Enviar Imagem", lá você vai encontrar tudo que precisa saber sobre envio neste formato.

Response#
201#
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

Atualizar nome do canal
Método#
/update-newsletter-name#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-newsletter-name

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por alterar o nome de um canal já existente.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
name	string	Novo nome do canal
Request Body#
URL#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-newsletter-name

Body#
{
  "id": "999999999999999999@newsletter",
  "name": "Novo nome do canal"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
201#
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

Atualizar descrição do canal
Método#
/update-newsletter-description#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-newsletter-description

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por alterar a descrição de um canal já existente.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
description	string	Nova descrição do canal
Request Body#
URL#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-newsletter-description

Body#
{
  "id": "999999999999999999@newsletter",
  "description": "Nova descrição do canal"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
201#
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

Seguir canal
Método#
/follow-newsletter#
PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/follow-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por seguir um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
Request Body#
URL#
Método

PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/follow-newsletter

Body#
{
  "id": "999999999999999999@newsletter"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST, PUT ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Deixar de seguir canal
Método#
/unfollow-newsletter#
PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/unfollow-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por deixar de seguir um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
Request Body#
URL#
Método

PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/unfollow-newsletter

Body#
{
  "id": "999999999999999999@newsletter"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST, PUT ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Mutar canal
Método#
/mute-newsletter#
PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/mute-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por mutar um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
Request Body#
URL#
Método

PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/mute-newsletter

Body#
{
  "id": "999999999999999999@newsletter"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST, PUT ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Desmutar canal
Método#
/unmute-newsletter#
PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/unmute-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por desmutar um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
Request Body#
URL#
Método

PUT https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/unmute-newsletter

Body#
{
  "id": "999999999999999999@newsletter"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST, PUT ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Deletar canal
Método#
/delete-newsletter#
DELETE https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/delete-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por deletar um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
Request Body#
URL#
Método

DELETE https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/delete-newsletter

Body#
{
  "id": "999999999999999999@newsletter"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
200#
Atributos	Tipo	Descrição
value	boolean	true caso tenha dado certo e false em caso de falha
Exemplo

{
  "value": true
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST, DELETE ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Metadata do canal
Método#
/newsletter/metadata#
GET https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/metadata/{newsletterId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna o metadata do canal com todas as informações do canal e de sua visualização.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal
Request Params#
URL#
GET https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/metadata/{newsletterId}

warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
200#
Atributos	Tipo	Descrição
id	string	ID do canal
creationTime	timestamp	Timestamp da data de criação do grupo
state	string	Estado do canal (ACTIVE, NON_EXISTING)
name	string	Nome do canal
description	string	Descrição do canal
subscribersCount	string	Contagem do número de seguidores do canal
inviteLink	string	Link de convite do canal
verification	string	Indica se o canal é verificado ou não (VERIFIED, UNVERIFIED)
picture	string	Url da imagem do canal
preview	string	Url de preview da imagem do canal
viewMetadata	object	Objeto com informações de visualização
Object (viewMetadata)

Atributos	Tipo	Descrição
mute	string	Indica se o canal esta mutado ou não (ON, OFF)
role	string	Indica se é o proprietário ou seguidor do canal (OWNER, SUBSCRIBER)
Exemplo

  {
    "id": "999999999999999999@newsletter",
    "creationTime": "1695643504",
    "state": "ACTIVE",
    "name": "Z-API",
    "description": "Canal oficial Z-API",
    "subscribersCount": "123",
    "inviteLink": "https://www.whatsapp.com/channel/0029Va5Xk71a",
    "verification": "VERIFIED",
    "picture": "https://mmg.whatsapp.net/v/t61.24694-24/383686038_859672472421500_990610487096734362_n.jpg?ccb=11-4&oh=01_AdS-Wk3RSfXmtEqDA4-LTFaZQILXZSprywV8EwNoZPOaGw&oe=651EF162&_nc_sid=000000&_nc_cat=111",
    "preview": "https://mmg.whatsapp.net/v/t61.24694-24/383686038_859672472421500_990610487096734362_n.jpg?stp=dst-jpg_s192x192&ccb=11-4&oh=01_AdRltWYOZftf0cnm-GNw5RRGoxQ53nJR9zzxxot_N7JQCw&oe=651EF162&_nc_sid=000000&_nc_cat=111",
    "viewMetadata": {
      "mute": "OFF",
      "role": "OWNER"
    }
  }

405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Listar canais
Método#
/newsletter#
GET https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna uma lista com o metadata dos canais próprios e seguidos com todas as informações do canal e de sua visualização.

URL#
GET https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter

Response#
200#
Atributos	Tipo	Descrição
id	string	ID do canal
creationTime	timestamp	Timestamp da data de criação do grupo
state	string	Estado do canal (ACTIVE, NON_EXISTING)
name	string	Nome do canal
description	string	Descrição do canal
subscribersCount	string	Contagem do número de seguidores do canal
inviteLink	string	Link de convite do canal
verification	string	Indica se o canal é verificado ou não (VERIFIED, UNVERIFIED)
picture	string	Url da imagem do canal
preview	string	Url de preview da imagem do canal
viewMetadata	object	Objeto com informações de visualização
Object (viewMetadata)

Atributos	Tipo	Descrição
mute	string	Indica se o canal esta mutado ou não (ON, OFF)
role	string	Indica se é o proprietário ou seguidor do canal (OWNER, SUBSCRIBER)
Exemplo

  [
    {
      "id": "999999999999999999@newsletter",
      "creationTime": "1695643504",
      "state": "ACTIVE",
      "name": "Z-API",
      "description": "Canal oficial Z-API",
      "subscribersCount": "123",
      "inviteLink": "https://www.whatsapp.com/channel/0029Va5Xk71a",
      "verification": "VERIFIED",
      "picture": "https://mmg.whatsapp.net/v/t61.24694-24/383686038_859672472421500_990610487096734362_n.jpg?ccb=11-4&oh=01_AdS-Wk3RSfXmtEqDA4-LTFaZQILXZSprywV8EwNoZPOaGw&oe=651EF162&_nc_sid=000000&_nc_cat=111",
      "preview": "https://mmg.whatsapp.net/v/t61.24694-24/383686038_859672472421500_990610487096734362_n.jpg?stp=dst-jpg_s192x192&ccb=11-4&oh=01_AdRltWYOZftf0cnm-GNw5RRGoxQ53nJR9zzxxot_N7JQCw&oe=651EF162&_nc_sid=000000&_nc_cat=111",
      "viewMetadata": {
        "mute": "OFF",
        "role": "OWNER"
      }
    },
    {
      "id": "999999999999999999@newsletter",
      "creationTime": "1695237295",
      "state": "ACTIVE",
      "name": "Canal Exemplo",
      "description": "Exemplo",
      "inviteLink": "https://www.whatsapp.com/channel/0029Va5Xk71a123",
      "verification": "UNVERIFIED",
      "picture": null,
      "preview": null,
      "viewMetadata": {
        "mute": "ON",
        "role": "SUBSCRIBER"
      }
    }
  ]
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Encontrar canais
Método#
/search-newsletter#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/search-newsletter

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método retorna uma lista com dados de canais, de acordo com a busca realizada através de filtros passados no body da requisição.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
limit	number	Limit de registros a serem listados
filters	object	Objeto com filtros a serem aplicados
Object (filters)

Atributos	Tipo	Descrição
countryCodes	array string	Array com codigo de países (https://www.iban.com/country-codes)
Opcionais#
Atributos	Tipo	Descrição
view	string	Filtro de visualização (RECOMMENDED, TRENDING, POPULAR, NEW)
searchText	string	Filtragem por texto
Request Body#
URL#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/search-newsletter

Body#
  {
    "limit": 50,
    "view": "TRENDING",
    "filters": {
      "countryCodes": ["BR", "AF", "CA"]
    },
    "searchText": "Z-API"
  }
Response#
200#
Atributos	Tipo	Descrição
id	string	ID do canal
name	string	Nome do canal
description	string	Descrição do canal
subscribersCount	string	Contagem do número de seguidores do canal
picture	string	Url da imagem do canal
Exemplo

  {
    "cursor": null,
    "data": [
      {
        "id": "999999999999999999@newsletter",
        "name": "Z-API",
        "description": "Canal oficial Z-API",
        "subscribersCount": "123",
        "picture": "https://mmg.whatsapp.net/v/t61.24694-24/345237462_968463277797373_5339431038113115975_n.jpg?stp=dst-jpg_s192x192&ccb=11-4&oh=01_AdTMyhA5kdwCdSqV0v784czJ1dHP_nkNhJ8TdgnANHro7Q&oe=651E6909&_nc_sid=000000&_nc_cat=109"
      },
      {
        "id": "999999999999999999@newsletter",
        "name": "Canal Exemplo",
        "description": "Exemplo",
        "subscribersCount": "0",
        "picture": null
      }
    ]
  }
Atributo "cursor" no objeto de resposta
A api do WhatsApp fornece o atributo "limit" para realizar a busca dos canais, o que significa que existe paginação dos resultados. Porém, na resposta não existe a indicação do "cursor" dos registros. Sendo assim, por enquanto, o atributo "cursor" sempre será "null", até que o WhatsApp implemente essa funcionalidade.

405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Atualizar configurações do canal
Método#
/newsletter/settings/{newsletterId}#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/settings/{newsletterId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por alterar as configurações de um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
id	string	ID do canal. Enviado no PATH da requisição (EX: newsletter/settings/999999999999999999@newsletter)
reactionCodes	string	Define a restrição de reações nas mensagens (basic, all)
(string) reactionCodes

Valores	Tipo	Descrição
basic	string	Permite apenas o envio de reações básicas
all	string	Permite o envio de qualquer reação
Request Body#
URL#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/settings/999999999999999999@newsletter

Body#
{
  "reactionCodes": "basic | all"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

Response#
201#
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

Aceitar convite de admin do canal
Método#
/newsletter/accept-admin-invite/{newsletterId}#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/accept-admin-invite/{newsletterId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por aceitar um convite para ser administrador de um canal. Esse convite, é uma mensagem que você tanto pode enviar quanto receber através do webhook de mensagem recebida.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
newsletterId	string	Id do canal o qual pertence o convite
Request#
Exemplo

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/accept-admin-invite/120363166555745933@newsletter

Response#
201#
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Remover admin do canal
Método#
/newsletter/remove-admin/{newsletterId}#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/remove-admin/{newsletterId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por remover um usuário da administração do canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
phone	string	Telefone do usuário que será removido da administração do canal
Request Body#

{
  "phone": "5511999999999"
}
Response#
201#
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Anular convite de admin do canal
Método#
/newsletter/revoke-admin-invite/{newsletterId}#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/revoke-admin-invite/{newsletterId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por anular um convite de administrador de um canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
phone	string	Telefone do usuário em que o convite será anulado
Request Body#

{
  "phone": "5511999999999"
}
Response#
201#
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"

Enviar convite de admin do canal
Método#
/send-newsletter-admin-invite#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/send-newsletter-admin-invite

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por enviar um convite para que um usuário se torne administrador de um canal. O convite é enviado como uma mensagem no WhatsApp para o número especificado.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
newsletterId	string	ID do canal
phone	string	Telefone do usuário que receberá o convite (formato internacional sem +)
Opcionais#
Atributos	Tipo	Descrição
comment	string	Comentário/mensagem adicional junto ao convite
Request Body#
URL#
Método

POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/send-newsletter-admin-invite

Body#
{
  "newsletterId": "999999999999999999@newsletter",
  "phone": "5511999999999"
}
Com comentário:

{
  "newsletterId": "999999999999999999@newsletter",
  "phone": "5511999999999",
  "comment": "Gostaria de convidá-lo para ser administrador do nosso canal"
}
warning
O id do canal sempre deve conter o sufixo "@newsletter", pois esse é o padrão utilizado pelo próprio WhatsApp.

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

Transferir propriedade do canal
Método#
/newsletter/transfer-ownership/{newsletterId}#
POST https://api.z-api.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/newsletter/transfer-ownership/{newsletterId}

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Conceituação#
Este método é responsável por transferir a propriedade de um canal a outro usuário, o qual seja administrador desse canal.

Atributos#
Obrigatórios#
Atributos	Tipo	Descrição
phone	string	Telefone do usuário que será promovido a dono do canal
Opcionais#
Atributos	Tipo	Descrição
quitAdmin	boolean	Define se você deixará de ser administrador do canal após transferir a propriedade
Request Body#
{
  "phone": "5511999999999"
}
{
  "phone": "5511999999999",
  "quitAdmin": true
}
Response#
200#
Atributos	Tipo	Descrição
value	string	Retorna true em caso de sucesso e false em caso de falha
message	string	Em caso de erro, pode retornar uma mensagem com informações sobre o erro
Exemplo

  {
    "value": true
  }
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou GET conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
