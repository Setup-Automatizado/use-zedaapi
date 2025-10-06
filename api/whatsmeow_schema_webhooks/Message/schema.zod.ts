import { z } from 'zod';

/**
* Schema Zod para evento Message
* Gerado automaticamente pelo Webhook Mapper
*/
export const MessageSchema = z.object({
  headers: z.object({
    host: z.string(),
    "user-agent": z.string(),
    "content-length": z.string(),
    accept: z.string(),
    "content-type": z.string(),
    "accept-encoding": z.string()
  }),
  body: z.object({
    baseURL: z.string(),
    data: z.object({
      event: z.object({
        Info: z.object({
          Chat: z.string(),
          Sender: z.string(),
          IsFromMe: z.boolean(),
          IsGroup: z.boolean(),
          AddressingMode: z.string(),
          SenderAlt: z.string(),
          RecipientAlt: z.string(),
          BroadcastListOwner: z.string(),
          ID: z.string(),
          ServerID: z.number(),
          Type: z.string(),
          PushName: z.string(),
          Timestamp: z.string(),
          Category: z.string(),
          Multicast: z.boolean(),
          MediaType: z.string(),
          Edit: z.string(),
          MsgBotInfo: z.object({
            EditType: z.string(),
            EditTargetID: z.string(),
            EditSenderTimestampMS: z.string()
          }),
          MsgMetaInfo: z.object({
            TargetID: z.string(),
            TargetSender: z.string(),
            TargetChat: z.string(),
            DeprecatedLIDSession: z.null(),
            ThreadMessageID: z.string(),
            ThreadMessageSenderJID: z.string()
          }),
          VerifiedName: z.union([z.object({}), z.null()]),
          DeviceSentMeta: z.union([z.null(), z.object({})])
        }),
        Message: z.object({
          conversation: z.string().optional(),
          messageContextInfo: z.object({
            messageSecret: z.string().optional(),
            deviceListMetadata: z.object({
              senderKeyHash: z.string(),
              senderTimestamp: z.number(),
              recipientKeyHash: z.string(),
              recipientTimestamp: z.number(),
              senderAccountType: z.number().optional(),
              receiverAccountType: z.number().optional()
            }).optional(),
            deviceListMetadataVersion: z.number().optional(),
            limitSharingV2: z.object({
              sharingLimited: z.boolean(),
              trigger: z.number(),
              limitSharingSettingTimestamp: z.number(),
              initiatedByMe: z.boolean()
            }).optional()
          }).optional(),
          senderKeyDistributionMessage: z.object({
            groupID: z.string(),
            axolotlSenderKeyDistributionMessage: z.string()
          }).optional(),
          imageMessage: z.object({
            URL: z.string(),
            mimetype: z.string(),
            caption: z.string().optional(),
            fileSHA256: z.string(),
            fileLength: z.number(),
            height: z.number(),
            width: z.number(),
            mediaKey: z.string(),
            fileEncSHA256: z.string(),
            directPath: z.string(),
            mediaKeyTimestamp: z.number(),
            /** TRUNCATED FIELD - Original: base64 */,
            JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
            contextInfo: z.object({
              expiration: z.number(),
              ephemeralSettingTimestamp: z.number().optional(),
              disappearingMode: z.object({
                initiator: z.number()
              }).optional()
            }).optional(),
            viewOnce: z.boolean().optional()
          }).optional(),
          extendedTextMessage: z.object({
            text: z.string(),
            previewType: z.number().optional(),
            contextInfo: z.object({
              stanzaID: z.string().optional(),
              participant: z.string().optional(),
              quotedMessage: z.object({
                conversation: z.string().optional(),
                imageMessage: z.object({
                  URL: z.string(),
                  mimetype: z.string(),
                  caption: z.string(),
                  fileSHA256: z.string(),
                  fileLength: z.number(),
                  height: z.number(),
                  width: z.number(),
                  mediaKey: z.string(),
                  fileEncSHA256: z.string(),
                  directPath: z.string(),
                  mediaKeyTimestamp: z.number(),
                  /** TRUNCATED FIELD - Original: base64 */,
                  JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
                  contextInfo: z.object({
                    pairedMediaType: z.number()
                  })
                }).optional(),
                messageContextInfo: z.object({
                  messageSecret: z.string(),
                  limitSharingV2: z.object({
                    sharingLimited: z.boolean(),
                    trigger: z.number(),
                    limitSharingSettingTimestamp: z.number(),
                    initiatedByMe: z.boolean()
                  }).optional()
                }).optional()
              }).optional(),
              expiration: z.number().optional(),
              ephemeralSettingTimestamp: z.number().optional(),
              disappearingMode: z.object({
                initiator: z.number(),
                trigger: z.number().optional()
              }).optional()
            }),
            inviteLinkGroupTypeV2: z.number().optional()
          }).optional(),
          pollUpdateMessage: z.object({
            pollCreationMessageKey: z.object({
              remoteJID: z.string(),
              fromMe: z.boolean(),
              ID: z.string(),
              participant: z.string()
            }),
            vote: z.object({
              encPayload: z.string(),
              encIV: z.string()
            }),
            senderTimestampMS: z.number()
          }).optional(),
          protocolMessage: z.object({
            key: z.object({
              remoteJID: z.string(),
              fromMe: z.boolean(),
              ID: z.string()
            }),
            type: z.number()
          }).optional(),
          videoMessage: z.object({
            URL: z.string(),
            mimetype: z.string(),
            fileSHA256: z.string(),
            fileLength: z.number(),
            seconds: z.number(),
            mediaKey: z.string(),
            height: z.number(),
            width: z.number(),
            fileEncSHA256: z.string(),
            directPath: z.string(),
            mediaKeyTimestamp: z.number(),
            /** TRUNCATED FIELD - Original: base64 */,
            JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
            contextInfo: z.object({
              pairedMediaType: z.number()
            }),
            streamingSidecar: z.string(),
            externalShareFullVideoDurationInSeconds: z.number()
          }).optional()
        }),
        IsEphemeral: z.boolean(),
        IsViewOnce: z.boolean(),
        IsViewOnceV2: z.boolean(),
        IsViewOnceV2Extension: z.boolean(),
        IsDocumentWithCaption: z.boolean(),
        IsLottieSticker: z.boolean(),
        IsBotInvoke: z.boolean(),
        IsEdit: z.boolean(),
        SourceWebMsg: z.null(),
        UnavailableRequestID: z.string(),
        RetryCount: z.number(),
        NewsletterMeta: z.null(),
        RawMessage: z.object({
          conversation: z.string().optional(),
          messageContextInfo: z.object({
            messageSecret: z.string().optional(),
            deviceListMetadata: z.object({
              senderKeyHash: z.string(),
              senderTimestamp: z.number(),
              recipientKeyHash: z.string(),
              recipientTimestamp: z.number(),
              senderAccountType: z.number().optional(),
              receiverAccountType: z.number().optional()
            }).optional(),
            deviceListMetadataVersion: z.number().optional(),
            limitSharingV2: z.object({
              sharingLimited: z.boolean(),
              trigger: z.number(),
              limitSharingSettingTimestamp: z.number(),
              initiatedByMe: z.boolean()
            }).optional()
          }).optional(),
          senderKeyDistributionMessage: z.object({
            groupID: z.string(),
            axolotlSenderKeyDistributionMessage: z.string()
          }).optional(),
          imageMessage: z.object({
            URL: z.string(),
            mimetype: z.string(),
            caption: z.string().optional(),
            fileSHA256: z.string(),
            fileLength: z.number(),
            height: z.number(),
            width: z.number(),
            mediaKey: z.string(),
            fileEncSHA256: z.string(),
            directPath: z.string(),
            mediaKeyTimestamp: z.number(),
            /** TRUNCATED FIELD - Original: base64 */,
            JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
            contextInfo: z.object({
              expiration: z.number(),
              ephemeralSettingTimestamp: z.number().optional(),
              disappearingMode: z.object({
                initiator: z.number()
              }).optional()
            }).optional(),
            viewOnce: z.boolean().optional()
          }).optional(),
          extendedTextMessage: z.object({
            text: z.string(),
            previewType: z.number().optional(),
            contextInfo: z.object({
              stanzaID: z.string().optional(),
              participant: z.string().optional(),
              quotedMessage: z.object({
                conversation: z.string().optional(),
                messageContextInfo: z.object({
                  messageSecret: z.string(),
                  limitSharingV2: z.object({
                    sharingLimited: z.boolean(),
                    trigger: z.number(),
                    limitSharingSettingTimestamp: z.number(),
                    initiatedByMe: z.boolean()
                  }).optional()
                }).optional(),
                imageMessage: z.object({
                  URL: z.string(),
                  mimetype: z.string(),
                  caption: z.string(),
                  fileSHA256: z.string(),
                  fileLength: z.number(),
                  height: z.number(),
                  width: z.number(),
                  mediaKey: z.string(),
                  fileEncSHA256: z.string(),
                  directPath: z.string(),
                  mediaKeyTimestamp: z.number(),
                  /** TRUNCATED FIELD - Original: base64 */,
                  JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
                  contextInfo: z.object({
                    pairedMediaType: z.number()
                  })
                }).optional()
              }).optional(),
              expiration: z.number().optional(),
              ephemeralSettingTimestamp: z.number().optional(),
              disappearingMode: z.object({
                initiator: z.number(),
                trigger: z.number().optional()
              }).optional()
            }),
            inviteLinkGroupTypeV2: z.number().optional()
          }).optional(),
          deviceSentMessage: z.object({
            destinationJID: z.string(),
            message: z.object({
              conversation: z.string().optional(),
              messageContextInfo: z.object({
                deviceListMetadata: z.object({
                  senderKeyHash: z.string(),
                  senderTimestamp: z.number(),
                  recipientKeyHash: z.string(),
                  recipientTimestamp: z.number(),
                  receiverAccountType: z.number().optional()
                }),
                deviceListMetadataVersion: z.number(),
                messageSecret: z.string()
              }),
              extendedTextMessage: z.object({
                text: z.string(),
                previewType: z.number(),
                contextInfo: z.object({
                  stanzaID: z.string(),
                  participant: z.string(),
                  quotedMessage: z.object({
                    imageMessage: z.object({
                      URL: z.string(),
                      mimetype: z.string(),
                      caption: z.string(),
                      fileSHA256: z.string(),
                      fileLength: z.number(),
                      height: z.number(),
                      width: z.number(),
                      mediaKey: z.string(),
                      fileEncSHA256: z.string(),
                      directPath: z.string(),
                      mediaKeyTimestamp: z.number(),
                      /** TRUNCATED FIELD - Original: base64 */,
                      JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
                      contextInfo: z.object({
                        pairedMediaType: z.number()
                      })
                    })
                  })
                }),
                inviteLinkGroupTypeV2: z.number()
              }).optional(),
              videoMessage: z.object({
                URL: z.string(),
                mimetype: z.string(),
                fileSHA256: z.string(),
                fileLength: z.number(),
                seconds: z.number(),
                mediaKey: z.string(),
                height: z.number(),
                width: z.number(),
                fileEncSHA256: z.string(),
                directPath: z.string(),
                mediaKeyTimestamp: z.number(),
                /** TRUNCATED FIELD - Original: base64 */,
                JPEGThumbnail: z.string().describe('TRUNCATED FIELD - Original type: base64'),
                contextInfo: z.object({
                  pairedMediaType: z.number()
                }),
                streamingSidecar: z.string(),
                externalShareFullVideoDurationInSeconds: z.number()
              }).optional()
            })
          }).optional(),
          pollUpdateMessage: z.object({
            pollCreationMessageKey: z.object({
              remoteJID: z.string(),
              fromMe: z.boolean(),
              ID: z.string(),
              participant: z.string()
            }),
            vote: z.object({
              encPayload: z.string(),
              encIV: z.string()
            }),
            senderTimestampMS: z.number()
          }).optional(),
          protocolMessage: z.object({
            key: z.object({
              remoteJID: z.string(),
              fromMe: z.boolean(),
              ID: z.string()
            }),
            type: z.number()
          }).optional()
        })
      }),
      type: z.string(),
      /** TRUNCATED FIELD - Original: base64 */,
      base64: z.string().describe('TRUNCATED FIELD - Original type: base64').optional(),
      fileName: z.string().optional(),
      mimeType: z.string().optional(),
      s3: z.object({
        bucket: z.string(),
        fileName: z.string(),
        key: z.string(),
        mimeType: z.string(),
        size: z.number(),
        url: z.string()
      }).optional()
    }),
    eventType: z.string(),
    timestamp: z.number(),
    token: z.string(),
    userID: z.string(),
    userJID: z.string(),
    userName: z.string()
  })
});

export type Message = z.infer<typeof MessageSchema>;
