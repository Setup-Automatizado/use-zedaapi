/**
 * Interface TypeScript para evento Mute
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Mute {
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
          muted?: boolean;
          muteEndTimestamp?: number;
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
