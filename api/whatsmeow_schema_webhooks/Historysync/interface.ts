/**
 * Interface TypeScript para evento Historysync
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Historysync {
  headers?: {
    host?: string;
    'user-agent'?: string;
    'content-length'?: string;
    accept?: string;
    'content-type'?: string;
    'accept-encoding'?: string;
  };
  body?: {
    data?: {
      event?: {
        Data?: {
          syncType?: number;
          conversations?: {
            ID?: string;
            messages?: {
              message?: {
                key?: {
                  remoteJID?: string;
                  fromMe?: boolean;
                  ID?: string;
                };
                message?: {
                  conversation?: string;
                  messageContextInfo?: {
                    messageSecret?: string;
                  };
                };
                messageTimestamp?: number;
                participant?: string;
                messageSecret?: string;
                isMentionedInStatus?: boolean;
              };
              msgOrderID?: number;
            }[];
            unreadCount?: number;
            readOnly?: boolean;
            ephemeralExpiration?: number;
            ephemeralSettingTimestamp?: number;
            endOfHistoryTransferType?: number;
            conversationTimestamp?: number;
            name?: string;
            pHash?: string;
            notSpam?: boolean;
            archived?: boolean;
            unreadMentionCount?: number;
            markedAsUnread?: boolean;
            suspended?: boolean;
          }[];
          chunkOrder?: number;
          progress?: number;
          phoneNumberToLidMappings?: {
            pnJID?: string;
            lidJID?: string;
          }[];
        };
      };
      type?: string;
    };
    eventType?: string;
    timestamp?: number;
    token?: string;
    userID?: string;
    userJID?: string;
    userName?: string;
  };
}
