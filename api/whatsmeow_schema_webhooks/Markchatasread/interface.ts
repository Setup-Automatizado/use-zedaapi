/**
 * Interface TypeScript para evento Markchatasread
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Markchatasread {
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
          read?: boolean;
          messageRange?: {};
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
