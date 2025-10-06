/**
 * Interface TypeScript para evento Pin
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Pin {
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
          pinned?: boolean;
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
