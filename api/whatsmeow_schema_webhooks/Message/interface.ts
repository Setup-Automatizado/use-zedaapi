/**
 * Interface TypeScript para evento Message
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Message {
  headers?: {
    host?: string;
    'user-agent'?: string;
    'content-length'?: string;
    accept?: string;
    'content-type'?: string;
    'accept-encoding'?: string;
  };
  body?: {
    baseURL?: string;
    data?: {
      event?: {
        Info?: {
          Chat?: string;
          Sender?: string;
          IsFromMe?: boolean;
          IsGroup?: boolean;
          AddressingMode?: string;
          SenderAlt?: string;
          RecipientAlt?: string;
          BroadcastListOwner?: string;
          ID?: string;
          ServerID?: number;
          Type?: string;
          PushName?: string;
          Timestamp?: string;
          Category?: string;
          Multicast?: boolean;
          MediaType?: string;
          Edit?: string;
          MsgBotInfo?: {
            EditType?: string;
            EditTargetID?: string;
            EditSenderTimestampMS?: string;
          };
          MsgMetaInfo?: {
            TargetID?: string;
            TargetSender?: string;
            TargetChat?: string;
            DeprecatedLIDSession?: null;
            ThreadMessageID?: string;
            ThreadMessageSenderJID?: string;
          };
          VerifiedName?: object | null;
          DeviceSentMeta?: null | object;
        };
        Message?: {
          conversation?: string;
          messageContextInfo?: {
            messageSecret?: string;
            deviceListMetadata?: {
              senderKeyHash?: string;
              senderTimestamp?: number;
              recipientKeyHash?: string;
              recipientTimestamp?: number;
              senderAccountType?: number;
              receiverAccountType?: number;
            };
            deviceListMetadataVersion?: number;
            limitSharingV2?: {
              sharingLimited?: boolean;
              trigger?: number;
              limitSharingSettingTimestamp?: number;
              initiatedByMe?: boolean;
            };
          };
          senderKeyDistributionMessage?: {
            groupID?: string;
            axolotlSenderKeyDistributionMessage?: string;
          };
          imageMessage?: {
            URL?: string;
            mimetype?: string;
            caption?: string;
            fileSHA256?: string;
            fileLength?: number;
            height?: number;
            width?: number;
            mediaKey?: string;
            fileEncSHA256?: string;
            directPath?: string;
            mediaKeyTimestamp?: number;
            /** TRUNCATED FIELD - Original: base64 */
            JPEGThumbnail?: string;
            contextInfo?: {
              expiration?: number;
              ephemeralSettingTimestamp?: number;
              disappearingMode?: {
                initiator?: number;
              };
            };
            viewOnce?: boolean;
          };
          extendedTextMessage?: {
            text?: string;
            previewType?: number;
            contextInfo?: {
              stanzaID?: string;
              participant?: string;
              quotedMessage?: {
                conversation?: string;
                imageMessage?: {
                  URL?: string;
                  mimetype?: string;
                  caption?: string;
                  fileSHA256?: string;
                  fileLength?: number;
                  height?: number;
                  width?: number;
                  mediaKey?: string;
                  fileEncSHA256?: string;
                  directPath?: string;
                  mediaKeyTimestamp?: number;
                  /** TRUNCATED FIELD - Original: base64 */
                  JPEGThumbnail?: string;
                  contextInfo?: {
                    pairedMediaType?: number;
                  };
                };
                messageContextInfo?: {
                  messageSecret?: string;
                  limitSharingV2?: {
                    sharingLimited?: boolean;
                    trigger?: number;
                    limitSharingSettingTimestamp?: number;
                    initiatedByMe?: boolean;
                  };
                };
              };
              expiration?: number;
              ephemeralSettingTimestamp?: number;
              disappearingMode?: {
                initiator?: number;
                trigger?: number;
              };
            };
            inviteLinkGroupTypeV2?: number;
          };
          pollUpdateMessage?: {
            pollCreationMessageKey?: {
              remoteJID?: string;
              fromMe?: boolean;
              ID?: string;
              participant?: string;
            };
            vote?: {
              encPayload?: string;
              encIV?: string;
            };
            senderTimestampMS?: number;
          };
          protocolMessage?: {
            key?: {
              remoteJID?: string;
              fromMe?: boolean;
              ID?: string;
            };
            type?: number;
          };
          videoMessage?: {
            URL?: string;
            mimetype?: string;
            fileSHA256?: string;
            fileLength?: number;
            seconds?: number;
            mediaKey?: string;
            height?: number;
            width?: number;
            fileEncSHA256?: string;
            directPath?: string;
            mediaKeyTimestamp?: number;
            /** TRUNCATED FIELD - Original: base64 */
            JPEGThumbnail?: string;
            contextInfo?: {
              pairedMediaType?: number;
            };
            streamingSidecar?: string;
            externalShareFullVideoDurationInSeconds?: number;
          };
        };
        IsEphemeral?: boolean;
        IsViewOnce?: boolean;
        IsViewOnceV2?: boolean;
        IsViewOnceV2Extension?: boolean;
        IsDocumentWithCaption?: boolean;
        IsLottieSticker?: boolean;
        IsBotInvoke?: boolean;
        IsEdit?: boolean;
        SourceWebMsg?: null;
        UnavailableRequestID?: string;
        RetryCount?: number;
        NewsletterMeta?: null;
        RawMessage?: {
          conversation?: string;
          messageContextInfo?: {
            messageSecret?: string;
            deviceListMetadata?: {
              senderKeyHash?: string;
              senderTimestamp?: number;
              recipientKeyHash?: string;
              recipientTimestamp?: number;
              senderAccountType?: number;
              receiverAccountType?: number;
            };
            deviceListMetadataVersion?: number;
            limitSharingV2?: {
              sharingLimited?: boolean;
              trigger?: number;
              limitSharingSettingTimestamp?: number;
              initiatedByMe?: boolean;
            };
          };
          senderKeyDistributionMessage?: {
            groupID?: string;
            axolotlSenderKeyDistributionMessage?: string;
          };
          imageMessage?: {
            URL?: string;
            mimetype?: string;
            caption?: string;
            fileSHA256?: string;
            fileLength?: number;
            height?: number;
            width?: number;
            mediaKey?: string;
            fileEncSHA256?: string;
            directPath?: string;
            mediaKeyTimestamp?: number;
            /** TRUNCATED FIELD - Original: base64 */
            JPEGThumbnail?: string;
            contextInfo?: {
              expiration?: number;
              ephemeralSettingTimestamp?: number;
              disappearingMode?: {
                initiator?: number;
              };
            };
            viewOnce?: boolean;
          };
          extendedTextMessage?: {
            text?: string;
            previewType?: number;
            contextInfo?: {
              stanzaID?: string;
              participant?: string;
              quotedMessage?: {
                conversation?: string;
                messageContextInfo?: {
                  messageSecret?: string;
                  limitSharingV2?: {
                    sharingLimited?: boolean;
                    trigger?: number;
                    limitSharingSettingTimestamp?: number;
                    initiatedByMe?: boolean;
                  };
                };
                imageMessage?: {
                  URL?: string;
                  mimetype?: string;
                  caption?: string;
                  fileSHA256?: string;
                  fileLength?: number;
                  height?: number;
                  width?: number;
                  mediaKey?: string;
                  fileEncSHA256?: string;
                  directPath?: string;
                  mediaKeyTimestamp?: number;
                  /** TRUNCATED FIELD - Original: base64 */
                  JPEGThumbnail?: string;
                  contextInfo?: {
                    pairedMediaType?: number;
                  };
                };
              };
              expiration?: number;
              ephemeralSettingTimestamp?: number;
              disappearingMode?: {
                initiator?: number;
                trigger?: number;
              };
            };
            inviteLinkGroupTypeV2?: number;
          };
          deviceSentMessage?: {
            destinationJID?: string;
            message?: {
              conversation?: string;
              messageContextInfo?: {
                deviceListMetadata?: {
                  senderKeyHash?: string;
                  senderTimestamp?: number;
                  recipientKeyHash?: string;
                  recipientTimestamp?: number;
                  receiverAccountType?: number;
                };
                deviceListMetadataVersion?: number;
                messageSecret?: string;
              };
              extendedTextMessage?: {
                text?: string;
                previewType?: number;
                contextInfo?: {
                  stanzaID?: string;
                  participant?: string;
                  quotedMessage?: {
                    imageMessage?: {
                      URL?: string;
                      mimetype?: string;
                      caption?: string;
                      fileSHA256?: string;
                      fileLength?: number;
                      height?: number;
                      width?: number;
                      mediaKey?: string;
                      fileEncSHA256?: string;
                      directPath?: string;
                      mediaKeyTimestamp?: number;
                      /** TRUNCATED FIELD - Original: base64 */
                      JPEGThumbnail?: string;
                      contextInfo?: {
                        pairedMediaType?: number;
                      };
                    };
                  };
                };
                inviteLinkGroupTypeV2?: number;
              };
              videoMessage?: {
                URL?: string;
                mimetype?: string;
                fileSHA256?: string;
                fileLength?: number;
                seconds?: number;
                mediaKey?: string;
                height?: number;
                width?: number;
                fileEncSHA256?: string;
                directPath?: string;
                mediaKeyTimestamp?: number;
                /** TRUNCATED FIELD - Original: base64 */
                JPEGThumbnail?: string;
                contextInfo?: {
                  pairedMediaType?: number;
                };
                streamingSidecar?: string;
                externalShareFullVideoDurationInSeconds?: number;
              };
            };
          };
          pollUpdateMessage?: {
            pollCreationMessageKey?: {
              remoteJID?: string;
              fromMe?: boolean;
              ID?: string;
              participant?: string;
            };
            vote?: {
              encPayload?: string;
              encIV?: string;
            };
            senderTimestampMS?: number;
          };
          protocolMessage?: {
            key?: {
              remoteJID?: string;
              fromMe?: boolean;
              ID?: string;
            };
            type?: number;
          };
        };
      };
      type?: string;
      /** TRUNCATED FIELD - Original: base64 */
      base64?: string;
      fileName?: string;
      mimeType?: string;
      s3?: {
        bucket?: string;
        fileName?: string;
        key?: string;
        mimeType?: string;
        size?: number;
        url?: string;
      };
    };
    eventType?: string;
    timestamp?: number;
    token?: string;
    userID?: string;
    userJID?: string;
    userName?: string;
  };
}
