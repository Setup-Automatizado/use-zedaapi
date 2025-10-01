# ATENÇÃO TOTAL AQUI:
## VAMOS IMPLEMENTAR AOS POUCOS OS NOVOS EVENTOS QUE FOREM SENDO IDENTIFICADOS MAS O PADRAO INICIAL SERÁ ESSA ESTRUTURA AQUI:

Esses retornos devem ser modularizados porque o que vem da Whatsmeow vem totalmente diferentes e vamos precisar transformar nesses abaixo:

Nossos webhooks#
Delivery#
Responsavel por avisar você que sua mensagem foi entregue ao WhatsApp, mas isso não significa necessáriamente que seu contato a recebeu, para informações de recebimento e leitura você vai precisar observar o webhook de status.

Receive#
Este webhook será chamado toda vez que alguem interagir com seu número no whatsapp.

Status#
Este método vai lhe avisar de todas mudanças de status que sua mensagem sofrer, se ela for recebida, lida, respondida ou excluida, ou seja uma mesma mensagem pode passar por varios status, e ter o mesmo status mais de uma vez, que é o caso de respondida.

Disconnected#
Este webhook será chamado sempre que nosso serviço identificar alguma indisponibiidade na comunicação, seja do seu celular com o whatsapp ou mesmo da conexão entre seu celular e o FunnelChatAPI.

```bash
Ao enviar
Conceituação#
Esse é o webhook de retorno de mensagens enviadas

API#
/update-webhook-delivery#
PUT https://api.funnelchatapi.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-webhook-delivery

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/delivery"
}

Retornos dos webhooks#
Os possíveis retornos do webhook on-message-send estão cadastrado logo abaixo:

Response#
Atributos	Tipo	Descrição
phone	string	Número de telefone de destino da mensagem.
zaapId	string	Identificador da mensagem na conversa.
type	string	Tipo do evento da instância, nesse caso será "DeliveryCallback".
200#
{
  "phone": "554499999999",
  "zaapId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "type": "DeliveryCallback",
  "instanceId": "instance.id"
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou PUT conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
```

```bash
Ao receber
Conceituação#
Esse é o webhook de retorno de mensagens recebidas, também é executada quando a sua instância está configurada para notificar também mensagens enviadas por você mesmo.

API#
/update-webhook-received#
PUT https://api.funnelchatapi.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-webhook-received

Ou#
Também é possível atualizar a rota com a opção "enviadas por mim" habilitada.

/update-webhook-received-delivery#
PUT https://api.funnelchatapi.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-webhook-received-delivery

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/receive"
}

Retornos dos webhooks#
Os possíveis retornos do webhook on-message-received estão cadastrado logo abaixo:

Response#
Atributos	Tipo	Descrição
isStatusReply	boolean	Identifica se a mensagem recebida é uma resposta de status
senderLid	string	ID do contato no whatsapp
connectedPhone	string	Número de telefone conectado à API
waitingMessage	boolean	Identifica se a sua mensagem está com status de "Aguardando a mensagem"
isEdit	boolean	Identifica se a mensagem recebida foi editada
isGroup	boolean	Indica se o chat é um grupo
isNewsletter	boolean	Indica se o chat é um canal
phone	string	Número de telefone, ou do grupo que enviou a mensagem.
fromMe	boolean	Indica se a mensagem enviada partiu do número conectado a API
participantPhone	string	Número de telefone do membro do grupo que enviou a mensagem.
participantLid	string	ID do contado no whatsapp, do participante que enviou a mensagem dentro de um grupo
messageId	string	Idetificador da mensagem na conversa.
status	string	Status que a mensagem se encontra no momento do envio (PENDING, SENT, RECEIVED, READ ou PLAYED).
referenceMessageId	string	Referência a mensagem que foi respondida para o caso da mensagem recebida ser uma resposta a uma mensagem da conversa.
momment	integer	Momento em que a mensagem foi recebida ou do erro.
messageExpirationSeconds	integer	Tempo das mensagens temporárias
requestMethod	string	Identificador do método de solicitação de entrada (invite_link ou non_admin_add)
type	string	Tipo do evento da instância, nesse caso será "ReceivedCallBack".
photo	string	Url da foto do usuário que enviou a mensagem.
text.message	string	Texto da mensagem.
image.caption	string	Leganda da foto.
image.imageUrl	string	Url da foto.
image.thumbnailUrl	string	Url da thumbnail da foto.
image.mimeType	string	MimeType da imagem.
audio.mimeType	string	MimeType do áudio.
audio.audioUrl	string	Url do áudio.
video.caption	string	Legenda do vídeo.
video.videoUrl	string	Url do vídeo.
video.mimeType	string	MimeType do vídeo.
contact.displayName	string	Nome do contato.
contact.vCard	string	VCard contendo as informações do contato.
document.mimeType	string	MimeType do documento.
document.fileName	string	Nome do documento.
document.title	string	Título do documento.
document.pageCount	string	Número de páginas do documento.
document.thumbnailUrl	string	Url da thumbnail do documento.
document.documentUrl	string	Url do documento.
location.thumbnailUrl	string	Url da thumbnail da localização.
location.longitude	float	Longitude da localização.
location.latitude	float	Latitude da localização.
location.url	string	Url da localização.
location.name	string	Nome da localização.
location.address	string	Endereço da localização.
sticker.mimeType	string	MimeType do sticker.
sticker.stickerUrl	string	Url do sticker.

# ATENÇÃO TOTAL AQUI:
Esses retornos devem ser modularizados porque o que vem da Whatsmeow vem totalmente diferentes e vamos precisar transformar nesses abaixo:

200#
Exemplo de retorno de texto#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228638000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "text": {
    "message": "teste",
    "descritpion": "(opcional) em caso da mensagem possuir uma descrição inserida pelo WhatsApp",
    "title": "(opcional) em caso da mensagem possuir um título inserido pelo WhatsApp",
    "url": "(opcional) caso a mensagem possua um link ligado a ela. Exemplo: mensagem de catálogo possui um botão 'Ver catálogo'",
    "thumbnailUrl": "(opcional) caso a mensagem possua uma imagem de thumbnail ligada a ela. Exemplo: mensagem de convite de grupo possui a imagem do grupo"
  }
}
Exemplo de retorno de template de texto#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "702CC5F7E0A6BF4421",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1708457193876,
  "status": "RECEIVED",
  "chatName": "Test Number",
  "senderPhoto": null,
  "senderName": "5544999999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "hydratedTemplate": {
    "header": {},
    "message": "texto da mensagem",
    "footer": "rodapé da mensagem",
    "title": "título da mensagem",
    "templateId": "790118069824606",
    "hydratedButtons": []
  }
}
Exemplo de retorno de reação#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228955000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "reaction": {
    "value": "❤️",
    "time": 1651878681150,
    "reactionBy": "554499999999",
    "referencedMessage": {
      "messageId": "3EB0796DC6B777C0C7CD",
      "fromMe": true,
      "phone": "5544999999999",
      "participant": null
    }
  }
}
Exemplo de retorno de texto (Lista de Botão)#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1634645380000,
  "status": "RECEIVED",
  "chatName": "Nome",
  "senderPhoto": "https://",
  "senderName": "Nome",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "referenceMessageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "forwarded": false,
  "type": "ReceivedCallback",
  "buttonsResponseMessage": {
    "buttonId": "1",
    "message": "Ótimo"
  }
}
Exemplo de retorno de template de botão OTP#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "9D968A5FA2880508C4",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708455444850,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "hydratedTemplate": {
    "header": {},
    "message": "texto da mensagem",
    "footer": "",
    "title": "",
    "templateId": "",
    "hydratedButtons": [
      {
        "urlButton": {
          "displayText": "Copiar código",
          "url": "https://www.whatsapp.com/otp/code/?otp_type=COPY_CODE&code=otp123"
        },
        "index": 0
      }
    ]
  }
}
Exemplo de retorno de botão de chave pix#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "9D968A5FA2880508C4",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708455444850,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "pixKeyMessage": {
    "currency": "BRL",
    "referenceId": "4PXRAHSIRDA",
    "key": "pixkey",
    "keyType": "EVP",
    "merchantName": "Pix"
  }
}
Exemplo de retorno de botão com imagem#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "9D968A5FA2880508C4",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708455444850,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "buttonsMessage": {
    "imageUrl": "URL da imagem",
    "videoUrl": null,
    "message": "Texto da mensagem",
    "buttons": [
      {
        "buttonId": "1",
        "type": 1,
        "buttonText": {
          "displayText": "Texto do botão 1"
        }
      },
      {
        "buttonId": "2",
        "type": 1,
        "buttonText": {
          "displayText": "Texto do botão 2"
        }
      }
    ]
  }
}
Exemplo de retorno de botão com video#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "9D968A5FA2880508C4",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708455444850,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "buttonsMessage": {
    "imageUrl": null,
    "videoUrl": "URL do video",
    "message": "Texto da mensagem",
    "buttons": [
      {
        "buttonId": "1",
        "type": 1,
        "buttonText": {
          "displayText": "Texto do botão 1"
        }
      },
      {
        "buttonId": "2",
        "type": 1,
        "buttonText": {
          "displayText": "Texto do botão 2"
        }
      }
    ]
  }
}
Exemplo de retorno de texto (Lista de Opcão)#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1634645683000,
  "status": "RECEIVED",
  "chatName": "Nome",
  "senderPhoto": "https://",
  "senderName": "Nome",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "referenceMessageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "forwarded": false,
  "type": "ReceivedCallback",
  "listResponseMessage": {
    "message": "FunnelChatAPI Asas para sua imaginação",
    "title": "FunnelChatAPI",
    "selectedRowId": "1"
  }
}
Exemplo de retorno de carrosel#
{
  "isStatusReply": false,
  "chatLid": null,
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "554499999999",
  "fromMe": true,
  "momment": 1739368022130,
  "status": "SENT",
  "chatName": "Nome",
  "senderPhoto": null,
  "senderName": "Nome",
  "photo": "https://",
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": true,
  "carouselMessage": {
    "text": "Texto da mensagem",
    "cards": [
      {
        "header": {
          "image": {
            "imageUrl": "https://",
            "thumbnailUrl": "https://",
            "caption": "",
            "mimeType": "image/jpeg",
            "viewOnce": false,
            "width": 0,
            "height": 0
          }
        },
        "message": "Mensagem do cartão do carrosel",
        "footer": "",
        "title": "",
        "hydratedButtons": [
          {
            "index": 0,
            "urlButton": {
              "displayText": "Texto do botão",
              "url": "https://"
            }
          },
          {
            "index": 1,
            "quickReplyButton": { "displayText": "Texto do botão", "id": "2" }
          }
        ]
      },
      {
        "header": {
          "image": {
            "imageUrl": "https://",
            "thumbnailUrl": "https://",
            "caption": "",
            "mimeType": "image/jpeg",
            "viewOnce": false,
            "width": 0,
            "height": 0
          }
        },
        "message": "Mensagem do cartão do carrosel",
        "footer": "",
        "title": "",
        "hydratedButtons": [
          {
            "index": 0,
            "urlButton": {
              "displayText": "Texto do botão",
              "url": "https://"
            }
          },
          {
            "index": 1,
            "quickReplyButton": { "displayText": "Texto do botão", "id": "2" }
          }
        ]
      }
    ]
  }
}
Exemplo de retorno de texto vindos de anúncio#
{
    "isStatusReply": false,
    "senderLid": "81896604192873@lid",
    "connectedPhone": "554499999999",
    "waitingMessage": false,
    "isGroup": false,
    "isEdit": false,
    "isNewsletter": false,
    "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
    "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
    "phone": "5544999999999",
    "fromMe": false,
    "momment": NumberLong(1657209752000),
    "status": "RECEIVED",
    "chatName": "name",
    "senderPhoto": null,
    "senderName": "name",
    "photo": null,
    "broadcast": false,
    "externalAdReply": {
        "title": "Titulo",
        "body": "texto do anuncio",
        "mediaType": NumberInt(1),
        "thumbnailUrl": "https://",
        "sourceType": "ad",
        "sourceId": "23722824350495506",
        "ctwaClid": "Aff-niaAw7V94N8LGd79Vjr43TlJD4UnoBdpZJQ3LzABitbbG6wgKBSVOth4EN0IDr9glsKWjm2LBaFrJG3Nb0ILxP49ZtossVBNzlS8cFXBvv2ow7gNw",
        "sourceUrl": "https://",
        "containsAutoReply": false,
        "renderLargerThumbnail": true,
        "showAdAttribution": true
    },
    "messageExpirationSeconds": NumberInt(0),
    "forwarded": false,
    "type": "ReceivedCallback",
    "text": {
        "message": "mensagem recebida",
        "description": "texto do anuncio",
        "title": "titulo",
        "url": "https://",
        "thumbnailUrl": "https://"
    }
}
Exemplo de retorno de imagem#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228828000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "image": {
    "mimeType": "image/jpeg",
    "imageUrl": "https://",
    "thumbnailUrl": "https://",
    "caption": "",
    "thumbnailUrl": "https://",
    "width": 600,
    "height": 315,
    "viewOnce": true
  }
}
Exemplo de retorno de template de imagem#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "885FF934BF100D579E",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708454725028,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "hydratedTemplate": {
    "header": {
      "image": {
        "imageUrl": "https://example.jpeg",
        "thumbnailUrl": "https://example.jpeg",
        "caption": "",
        "mimeType": "image/jpeg",
        "viewOnce": false,
        "width": 1600,
        "height": 926
      }
    },
    "message": "texto da mensagem",
    "footer": "rodapé da mensagem",
    "title": "título da mensagem",
    "templateId": "674504507982622",
    "hydratedButtons": []
  }
}
Exemplo de retorno de áudio#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228849000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "audio": {
    "ptt": true,
    "seconds": 10,
    "audioUrl": "https://",
    "mimeType": "audio/ogg; codecs=opus",
    "viewOnce": true
  }
}
Exemplo de retorno de vídeo#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228889000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "video": {
    "videoUrl": "https://",
    "caption": "",
    "mimeType": "video/mp4",
    "seconds": 13,
    "viewOnce": true
  }
}
Exemplo de retorno de template de vídeo#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "0E4AD761B62E3D5EF9",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1708456287181,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "5544999999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "hydratedTemplate": {
    "header": {
      "video": {
        "videoUrl": "https://example.mp4",
        "caption": "",
        "mimeType": "video/mp4",
        "width": 0,
        "height": 0,
        "seconds": 0,
        "viewOnce": false
      }
    },
    "message": "texto da mensagem",
    "footer": "rodapé da mensagem",
    "title": "título da mensagem",
    "templateId": "938481574354947",
    "hydratedButtons": []
  }
}
Exemplo de retorno de PTV#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": true,
  "momment": 1688496074000,
  "status": "RECEIVED",
  "chatName": "eu",
  "senderPhoto": "https://",
  "senderName": "name",
  "photo": "https://",
  "broadcast": false,
  "participantPhone": "5544999999999",
  "messageExpirationSeconds": 0,
  "forwarded": true,
  "type": "ReceivedCallback",
  "video": {
    "videoUrl": "https://",
    "caption": "",
    "mimeType": "video/mp4"
  }
}
Exemplo de retorno de contato#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228925000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "contact": {
    "displayName": "Cesar Baleco",
    "vCard": "BEGIN:VCARD\nVERSION:3.0\nN:;nome;;;\nFN:nome\nTEL;type=CELL;type=VOICE;waid=5544999999999:+55 44 9999-9999\nEND:VCARD",
    "phones": ["5544999999999"]
  }
}
Exemplo de retorno de documento#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228955000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "document": {
    "documentUrl": "https://",
    "mimeType": "application/pdf",
    "title": "nome",
    "pageCount": 1,
    "fileName": "nome.pdf"
  }
}
Exemplo de retorno de template de documento#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "9D968A5FA2880508C4",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708455444850,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "hydratedTemplate": {
    "header": {
      "document": {
        "caption": null,
        "documentUrl": "https://example.pdf",
        "mimeType": "application/pdf",
        "title": "",
        "pageCount": 0,
        "fileName": ""
      }
    },
    "message": "texto da mensagem",
    "footer": "rodapé da mensagem",
    "title": "título da mensagem",
    "templateId": "811492407484976",
    "hydratedButtons": []
  }
}
Exemplo de retorno de localização#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228970000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "location": {
    "longitude": -99.999999999999999,
    "latitude": -99.9999999999999999,
    "address": "",
    "url": ""
  }
}
Exemplo de retorno de template de localização#
{
  "isStatusReply": false,
  "chatLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "3C67AB641C8AA0412F6A2242B4E23AC7",
  "messageId": "27BBF23E0185D363D9",
  "phone": "554499999999",
  "fromMe": false,
  "momment": 1708456969808,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "554499999999",
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "fromApi": false,
  "hydratedTemplate": {
    "header": {
      "location": {
        "longitude": -46.6388,
        "latitude": -23.5489,
        "name": "nome do lugar",
        "address": "nome do enderço",
        "url": ""
      }
    },
    "message": "texto da mensagem",
    "footer": "rodapé da mensagem",
    "title": "título da mensagem",
    "templateId": "1143940003434066",
    "hydratedButtons": []
  }
}
Exemplo de retorno de sticker#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228982000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "sticker": {
    "stickerUrl": "https://",
    "mimeType": "image/webp"
  }
}
Exemplo de retorno de GIF#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228889000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "video": {
    "videoUrl": "https://",
    "caption": "",
    "mimeType": "video/mp4"
  }
}
Exemplo de retorno de pagamento feito#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632229683000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "requestPayment": {
    "value": 1,
    "currencyCode": "BRL",
    "expiration": 1632834482000,
    "requestPhone": "5544999999999",
    "paymentInfo": {
      "receiverPhone": "5544999999999",
      "value": 1,
      "currencyCode": "BRL",
      "status": "WAITING",
      "transactionStatus": "COLLECT_SUCCESS"
    }
  }
}
Exemplo de retorno de pedido de pagamento#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": true,
  "momment": 1632230332000,
  "status": "MESSAGE_RECEIVED",
  "chatName": "name",
  "senderName": "name",
  "participantPhone": "5544999999999",
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "notification": "PAYMENT_ACTION_REQUEST_DECLINED",
  "notificationParameters": ["5544999999999", "BRL", "1000"]
}
Exemplo de retorno de recebimento de pagamento#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632230512000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "sendPayment": {
    "paymentInfo": {
      "receiverPhone": "5544999999999",
      "value": 1,
      "currencyCode": "BRL",
      "status": "COMPLETE",
      "transactionStatus": "SUCCESS"
    }
  }
}
Exemplo de retorno de ligação recebida#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isGroup": false,
  "isNewsletter": false,
  "isEdit": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "1679655074-84",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1679661190000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "photo": "https://",
  "broadcast": false,
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "CALL_VOICE",
  "notificationParameters": [],
  "callId": "F44E0E2011E7C784BB9A4AC11749C436"
}
Exemplo de retorno de ligação perdida#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "1679655074-103",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1679661194000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "",
  "photo": "https://",
  "broadcast": false,
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "CALL_MISSED_VOICE",
  "notificationParameters": [],
  "callId": "F44E0E2011E7C784BB9A4AC11749C436"
}
Exemplo de solicitação de entrada em grupo através de um link de convite#
{
  "isGroup": true,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999-group",
  "connectedPhone": "5544999999999",
  "fromMe": false,
  "momment": 1682017970000,
  "expiresAt": null,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "photo": null,
  "broadcast": false,
  "participantPhone": "5544999999999",
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "MEMBERSHIP_APPROVAL_REQUEST",
  "notificationParameters": [
      "5544999999999"
  ],
  "callId": null,
  "code": null,
  "requestMethod": "invite_link"
}
Exemplo de solicitação de entrada em grupo revogada pelo usuário#
{
  "isGroup": true,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999-group",
  "connectedPhone": "5544999999999",
  "fromMe": false,
  "momment": 1682017970000,
  "expiresAt": null,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "photo": null,
  "broadcast": false,
  "participantPhone": "5544999999999",
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "REVOKED_MEMBERSHIP_REQUESTS",
  "notificationParameters": [
      "5544999999999"
  ],
  "callId": null,
  "code": null
}
Exemplo de solicitação de entrada em grupo adicionado por um participante#
{
  "isGroup": true,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999-group",
  "connectedPhone": "5544999999999",
  "fromMe": false,
  "momment": 1682017970000,
  "expiresAt": null,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "photo": null,
  "broadcast": false,
  "participantPhone": "5544999999999",
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "MEMBERSHIP_APPROVAL_REQUEST",
  "notificationParameters": [
      "5544999999999",
      "5544888888888"
  ],
  "callId": null,
  "code": null,
  "requestMethod": "non_admin_add"
}
Exemplo de admin promovido a um canal#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "464201093",
  "phone": "5544999999999@newsletter",
  "fromMe": false,
  "momment": 1682017970000,
  "status": "RECEIVED",
  "chatName": "nome do canal",
  "senderPhoto": null,
  "senderName": "",
  "photo": null,
  "broadcast": false,
  "participantPhone": "5544999999999",
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "NEWSLETTER_ADMIN_PROMOTE",
  "notificationParameters": ["5544999999999", "ADMIN"],
  "callId": null
}
Exemplo de admin removido de um canal#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "464201093",
  "phone": "5544999999999@newsletter",
  "fromMe": false,
  "momment": 1682017970000,
  "status": "RECEIVED",
  "chatName": "nome do canal",
  "senderPhoto": null,
  "senderName": "",
  "photo": null,
  "broadcast": false,
  "participantPhone": "5544999999999",
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "NEWSLETTER_ADMIN_DEMOTE",
  "notificationParameters": ["5544999999999", "SUBSCRIBER"],
  "callId": null
}
Exemplo de retorno de produto#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632233527000,
  "status": "RECEIVED",
  "senderPhoto": "https://",
  "senderName": "5544999999999",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "product": {
    "productImage": "https://",
    "businessOwnerJid": "5544999999999",
    "currencyCode": "BRL",
    "productId": "99999999999999999999",
    "description": "",
    "productImageCount": 1,
    "price": 1,
    "url": "",
    "retailerId": "",
    "firstImageId": "",
    "title": "name"
  }
}
Exemplo de retorno de carrinho#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632233527000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": null,
  "senderName": "name",
  "photo": "https://",
  "broadcast": false,
  "forwarded": false,
  "type": "ReceivedCallback",
  "order": {
    "itemCount": 1,
    "orderId": "422508169684569",
    "message": "",
    "orderTitle": "name",
    "sellerJid": "5544999999999",
    "thumbnailUrl": "https://",
    "token": "AR5d4yUr+DmSzeCR2kUtPEeMfS+eG0O+S/T/17B+oY1mfA==",
    "currency": "BRL",
    "total": 2000,
    "subTotal": 2000,
    "products": [
      {
        "quantity": 1,
        "name": "nameProduct",
        "productId": "5338924696127051",
        "retailerId": "1242",
        "price": 2000,
        "currencyCode": "BRL"
      }
    ]
  }
}
Exemplo de retorno de enquete#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228638000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "poll": {
    "question": "Qual a melhor API de WhatsApp?",
    "pollMaxOptions": 0,
    "options": [
      {
        "name": "FunnelChatAPI"
      },
      {
        "name": "Outras"
      }
    ]
  }
}
Exemplo de retorno de resposta de enquete#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228638000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": "se for grupo esse será o participante que respondeu",
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "pollVote": {
    "pollMessageId": "ID da mensagem de enquete que foi respondida",
    "options": [
      {
        "name": "FunnelChatAPI"
      }
    ]
  }
}
Exemplo de retorno de envio de pedido#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228925000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "reviewAndPay": {
    "type": "physical-goods",
    "currency": "BRL",
    "referenceId": "4N9AVI38VOB",
    "orderRequestId": "4N9AVI38VYZ",
    "orderStatus": "pending",
    "paymentStatus": "pending",
    "total": 605,
    "subTotal": 600,
    "discount": 10,
    "shipping": 5,
    "tax": 10,
    "products": [
      {
        "name": "order 1",
        "quantity": 2,
        "isCustomItem": true,
        "productId": "custom-item-4N9AVI38WI1",
        "value": 150
      },
      {
        "name": "order 2",
        "quantity": 2,
        "isCustomItem": false,
        "productId": "23940797548900636",
        "value": 150
      }
    ]
  }
}
Exemplo de retorno de atualização de pedido#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228925000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "reviewOrder": {
    "currency": "BRL",
    "referenceId": "4N9AVI38VOB",
    "orderRequestId": "4N9AVI38VYZ",
    "orderStatus": "processing",
    "paymentStatus": "pending",
    "total": 605,
    "subTotal": 600,
    "discount": 10,
    "shipping": 5,
    "tax": 10,
    "products": [
      {
        "name": "order 1",
        "quantity": 2,
        "isCustomItem": true,
        "productId": "custom-item-4N9AVI38WI1",
        "value": 150
      },
      {
        "name": "order 2",
        "quantity": 2,
        "isCustomItem": false,
        "productId": "23940797548900636",
        "value": 150
      }
    ]
  }
}
Exemplo de retorno de convite admin de canal#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": false,
  "momment": 1632228925000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "newsletterAdminInvite": {
    "newsletterId": "120363166555745933@newsletter",
    "newsletterName": "Teste",
    "text": "Quero convidar vocÃª para ser admin do meu canal no WhatsApp.",
    "inviteExpiration": 1706809668
  }
}
Exemplo de retorno de fixar mensagem#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "fromMe": true,
  "momment": 1632228955000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "pinMessage": {
    "action": "pin",
    "pinDurationInSecs": 604800,
    "referencedMessage": {
      "messageId": "3EB0796DC6B777C0C7CD",
      "fromMe": true,
      "phone": "554499999999",
      "participant": null
    }
  }
}
Exemplo de retorno de evento#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": true,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "120363019502650977-group",
  "fromMe": false,
  "momment": 1632228638000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "event": {
    "name": "Nome do evento",
    "description": "Descrição do evento",
    "canceled": false,
    "joinLink": "https://call.whatsapp.com/video/v9123XNFG50L6iO79NddXNvKQr6bb3",
    "scheduleTime": 1716915653,
    "location": {}
  }
}
Exemplo de retorno de resposta de evento#
{
  "isStatusReply": false,
  "senderLid": "81896604192873@lid",
  "connectedPhone": "554499999999",
  "waitingMessage": false,
  "isEdit": false,
  "isGroup": true,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "120363019502650977-group",
  "fromMe": false,
  "momment": 1632228638000,
  "status": "RECEIVED",
  "chatName": "name",
  "senderPhoto": "https://",
  "senderName": "name",
  "participantPhone": null,
  "participantLid": null,
  "photo": "https://",
  "broadcast": false,
  "type": "ReceivedCallback",
  "eventResponse": {
    "response": "GOING",
    "responseFrom": "554499999999",
    "time": 1714423417000,
    "referencedMessage": {
      "messageId": "D2D612289D9E8F62307D72409A8D9DC3",
      "fromMe": false,
      "phone": "120363239161320697-group",
      "participant": "554499999988"
    }
  }
}
Exemplo de retorno de "aguardando mensagem"#
{
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "messageId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "momment": 1736797729000,
  "status": "RECEIVED",
  "fromMe": true,
  "phone": "5544999999999",
  "chatName": "chat",
  "senderName": "name",
  "senderPhoto": null,
  "photo": null,
  "broadcast": false,
  "participantLid": null,
  "type": "ReceivedCallback",
  "waitingMessage": true,
  "viewOnce": true
}
Exemplo de retorno de alteração do Nome do Whatsapp conectado#
{
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "connectedPhone": "5544999999999",
  "fromMe": true,
  "momment": 1736797729000,
  "expiresAt": null,
  "status": "RECEIVED",
  "chatName": null,
  "senderPhoto": "https://",
  "senderName": "nome",
  "photo": "https://",
  "broadcast": false,
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "PROFILE_NAME_UPDATED",
  "callId": null,
  "code": null,
  "profileName": "nome atualizado"
}
Exemplo de retorno de alteração da foto do Whatsapp conectado#
{
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "connectedPhone": "5544999999999",
  "fromMe": true,
  "momment": 1736797729000,
  "expiresAt": null,
  "status": "RECEIVED",
  "chatName": null,
  "senderPhoto": "https://",
  "senderName": "nome",
  "photo": "https://",
  "broadcast": false,
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "PROFILE_PICTURE_UPDATED",
  "callId": null,
  "code": null,
  "updatedPhoto": "https://"
}
Exemplo de retorno de alteração de etiquetas de um chat#
{
  "isGroup": false,
  "isNewsletter": false,
  "instanceId": "A20DA9C0183A2D35A260F53F5D2B9244",
  "phone": "5544999999999",
  "connectedPhone": "5544999999999",
  "fromMe": true,
  "momment": 1736797729000,
  "expiresAt": null,
  "status": "RECEIVED",
  "chatName": null,
  "senderPhoto": null,
  "senderName": "name",
  "photo": null,
  "broadcast": false,
  "referenceMessageId": null,
  "externalAdReply": null,
  "forwarded": false,
  "type": "ReceivedCallback",
  "notification": "CHAT_LABEL_ASSOCIATION",
  "notificationParameters": [
    {
      "phone": "5544977777777",
      "label": "1",
      "assigned": true
    },
    {
      "phone": "5544988888888",
      "label": "2",
      "assigned": false
    }
  ],
  "callId": null,
  "code": null
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou PUT conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
```

```bash
Ao desconectar
Conceituação#
Esse é o webhook de resposta de desconexão

API#
/update-webhook-disconnected#
PUT https://api.funnelchatapi.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-webhook-disconnected

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/disconnected"
}

Retornos dos webhooks#
Os possíveis retornos do webhook on-whatsapp-disconnected estão cadastrado logo abaixo:

Response#
Atributos	Tipo	Descrição
momment	integer	Momento em que a instância foi desconectada do número.
error	string	Descrição do erro.
disconnected	boolean	Indicação se a instância está conectada com o número ou não.
type	string	Tipo do evento da instância, nesse caso será "DisconnectedCallback".
200#
{
  "momment": 1580163342,
  "error": "Device has been disconnected",
  "disconnected": true,
  "type": "DisconnectedCallback",
  "instanceId": "instance.id"
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou PUT conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
```

```bash
Status da mensagem
Conceituação#
Esse é o webhook de retorno do status da mensagem

API#
/update-webhook-status#
PUT https://api.funnelchatapi.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-webhook-message-status

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/status"
}

Retornos dos webhooks#
Os possíveis retornos do webhook on-whatsapp-message-status-changes estão cadastrado logo abaixo:

Response#
Atributos	Tipo	Descrição
status	string	Status da mensagem (SENT - se foi enviada, RECEIVED - se foi recebida, READ - se foi lida, PLAYED - se foi ouvida )
id	string	Identificador(es) da(s) mensagem(ns).
momment	integer	Momento em que a instância foi desconectada do número.
phoneDevice	integer	Indica o dispositivo que ocorreu o evento (0 - Celular)
phone	string	Número de telefone de destino da mensagem.
type	string	Tipo do evento da instância, nesse caso será "MessageStatusCallback".
isGroup	boolean	Indica se o chat é um grupo

# ATENÇÃO TOTAL AQUI:
Esses retornos devem ser modularizados porque o que vem da Whatsmeow vem totalmente diferentes e vamos precisar transformar nesses abaixo:

200#
{
  "instanceId": "instance.id",
  "status": "SENT",
  "ids": ["999999999999999999999"],
  "momment": 1632234645000,
  "phoneDevice": 0,
  "phone": "5544999999999",
  "type": "MessageStatusCallback",
  "isGroup": false
}
{
  "instanceId": "instance.id",
  "status": "RECEIVED",
  "ids": ["999999999999999999999"],
  "momment": 1632234645000,
  "phoneDevice": 0,
  "phone": "5544999999999",
  "type": "MessageStatusCallback",
  "isGroup": false
}
{
  "instanceId": "instance.id",
  "status": "READ",
  "ids": ["999999999999999999999"],
  "momment": 1632234645000,
  "phoneDevice": 0,
  "phone": "5544999999999",
  "type": "MessageStatusCallback",
  "isGroup": false
}
{
  "instanceId": "instance.id",
  "status": "PLAYED",
  "ids": ["999999999999999999999"],
  "momment": 1632234645000,
  "phoneDevice": 0,
  "phone": "5544999999999",
  "type": "MessageStatusCallback",
  "isGroup": false
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou PUT conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
```

```bash
Status do chat
Conceituação#
Esse é o webhook de retorno status do chat

API#
/update-webhook-chat-presence#
PUT https://api.funnelchatapi.io/instances/SUA_INSTANCIA/token/SEU_TOKEN/update-webhook-chat-presence

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/presence"
}

Retornos dos webhooks#
Os possíveis retornos do webhook on-chat-presence estão cadastrado logo abaixo:

Response#
Atributos	Tipo	Descrição
type	string	Tipo do evento da instância, nesse caso será "PresenceChatCallback".
phone	string	Número de telefone de destino da mensagem.
status	string	Identificador do status do chat ex: (Digitando...) status pode conter ( UNAVAILABLE, AVAILABLE, COMPOSING, RECORDING)
lastSeen	timestamp	Identificador da ultima vez presente do usuário.

# ATENÇÃO TOTAL AQUI:
Esses retornos devem ser modularizados porque o que vem da Whatsmeow vem totalmente diferentes e vamos precisar transformar nesses abaixo:

200#
Fora do chat#
{
  "type": "PresenceChatCallback",
  "phone": "5544999999999",
  "status": "UNAVAILABLE",
  "lastSeen": null,
  "instanceId": "instance.id"
}
Dentro do chat#
{
  "type": "PresenceChatCallback",
  "phone": "5544999999999",
  "status": "AVAILABLE",
  "lastSeen": null,
  "instanceId": "instance.id"
}
Digitando no chat#
{
  "type": "PresenceChatCallback",
  "phone": "5544999999999",
  "status": "COMPOSING",
  "lastSeen": null,
  "instanceId": "instance.id"
}
Parou de digitar ou apagou o que estava digitando#
{
  "type": "PresenceChatCallback",
  "phone": "5544999999999",
  "status": "PAUSED",
  "lastSeen": null,
  "instanceId": "instance.id"
}
Gravando áudio no chat#
{
  "type": "PresenceChatCallback",
  "phone": "5544999999999",
  "status": "RECORDING",
  "lastSeen": null,
  "instanceId": "instance.id"
}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou PUT conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
```

```bash
Ao conectar
Conceituação#
Esse é o webhook de retorno de conexão do Celular com o FunnelChatAPI

Esse é webhook é disparado quando o FunnelChatAPI se conectada ao WhatsApp, isso pode acontecer quando le o qr-code, quando reiniciar a instancia e etc

API#
/update-webhook-connected#
PUT https://api.funnelchatapi.io/instances/{id}/token/{token}/update-webhook-connected

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/status"
}

Retornos dos webhooks#
Os possíveis retornos do webhook on-webhook-connected estão cadastrado logo abaixo:

Response#
Atributos	Tipo	Descrição
connected	boolean	status da instância.
phone	string	Número conectado.
momment	string	Momento em que a instância foi desconectada do número.
type	string	Tipo do evento da instância, nesse caso será "ConnectedCallback" .

# ATENÇÃO TOTAL AQUI:
Esses retornos devem ser modularizados porque o que vem da Whatsmeow vem totalmente diferentes e vamos precisar transformar nesses abaixo:

200#
{

  "type": 'ConnectedCallback',
  "connected": true,
  "momment": 26151515154,
  "instanceId": instance.id,
  "phone": "número",,
  "instanceId": "instance.id"

}
405#
Neste caso certifique que esteja enviando o corretamente a especificação do método, ou seja verifique se você enviou o POST ou PUT conforme especificado no inicio deste tópico.

415#
Caso você receba um erro 415, certifique de adicionar na headers da requisição o "Content-Type" do objeto que você está enviando, em sua grande maioria "application/json"
```

```bash
Atualizar todos webhooks
Conceituação#
Esse endpoint serve para você que deseja alterar todos os webhooks para a mesma URL de uma só vez.

API#
/update-every-webhooks#
PUT https://api.funnelchatapi.io/instances/{id}/token/{token}/update-every-webhooks

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
Obrigatórios#
Atributos	Tipo	Descrição
value	string	Endpoint do webhook
Opcionais#
Atributos	Tipo	Descrição
notifySentByMe	boolean	Ativar webhook de mensagens recebidas e enviadas por mim
{
  "value": "https://endereco-do-seu-sistema.com.br/instancia/SUA_INSTANCIA/status",
  "notifySentByMe": true
}
Retorno do endpoint#
200#
{
  "value": boolean
}
```

```bash
Atualizar notificar enviadas por mim
Conceituação#
Esse endpoint serve para você habilitar a opção de receber mensagens enviadas por você através do webhook.

API#
/update-notify-sent-by-me#
PUT https://api.z-api.io/instances/{id}/token/{token}/update-notify-sent-by-me

Header#
Key	Value
Client-Token	TOKEN DE SEGURANÇA DA CONTA
Request Body#
Obrigatórios#
Atributos	Tipo	Descrição
notifySentByMe	boolean	Ativar webhook de mensagens recebidas e enviadas por mim
{
  "notifySentByMe": true
}
Retorno do endpoint#
200#
{
    "value": boolean
}
```
