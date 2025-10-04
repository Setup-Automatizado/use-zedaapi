/**
 * Interface TypeScript para evento Archive
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Archive {
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
        JID?: string;
        Timestamp?: string;
        Action?: {
          archived?: boolean;
          messageRange?: {
            lastMessageTimestamp?: number;
            messages?: {
              key?: {
                remoteJID?: string;
                fromMe?: boolean;
                ID?: string;
              };
              timestamp?: number;
            }[];
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
  };
}
