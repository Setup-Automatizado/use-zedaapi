/**
 * Interface TypeScript para evento Blocklist
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Blocklist {
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
        Action?: string;
        DHash?: string;
        PrevDHash?: string;
        Changes?: {
          JID?: string;
          Action?: string;
        }[];
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
