/**
 * Interface TypeScript para evento Deletechat
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Deletechat {
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
        JID?: string;
        Timestamp?: string;
        Action?: {
          messageRange?: {
            lastMessageTimestamp?: number;
            messages?: {
              key?: {
                remoteJID?: string;
                fromMe?: boolean;
                ID?: string;
                participant?: string;
              };
              timestamp?: number;
            }[];
            lastSystemMessageTimestamp?: number;
          };
        };
        FromFullSync?: boolean;
      };
      type?: string;
    };
    eventType?: string;
    timestamp?: number;
    token?: string;
    userID?: string;
    userJID?: string;
    userName?: string;
    baseURL?: string;
  };
}
