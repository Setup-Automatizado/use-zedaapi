/**
 * Interface TypeScript para evento Picture
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Picture {
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
        Author?: string;
        Timestamp?: string;
        Remove?: boolean;
        PictureID?: string;
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
