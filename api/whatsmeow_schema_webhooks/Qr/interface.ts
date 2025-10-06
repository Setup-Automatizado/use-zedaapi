/**
 * Interface TypeScript para evento Qr
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Qr {
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
      code?: string;
      event?: string;
      /** TRUNCATED FIELD - Original: text */
      qrCodeBase64?: string;
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
