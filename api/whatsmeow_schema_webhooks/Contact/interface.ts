/**
 * Interface TypeScript para evento Contact
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Contact {
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
          fullName?: string;
          firstName?: string;
          lidJID?: string;
          saveOnPrimaryAddressbook?: boolean;
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
