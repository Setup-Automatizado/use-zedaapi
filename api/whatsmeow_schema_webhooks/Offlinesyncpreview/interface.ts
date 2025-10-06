/**
 * Interface TypeScript para evento Offlinesyncpreview
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Offlinesyncpreview {
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
        Total?: number;
        AppDataChanges?: number;
        Messages?: number;
        Notifications?: number;
        Receipts?: number;
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
