/**
 * Interface TypeScript para evento Deleteforme
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Deleteforme {
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
        ChatJID?: string;
        SenderJID?: string;
        IsFromMe?: boolean;
        MessageID?: string;
        Timestamp?: string;
        Action?: {
          deleteMedia?: boolean;
          messageTimestamp?: number;
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
