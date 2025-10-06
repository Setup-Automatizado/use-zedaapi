/**
 * Interface TypeScript para evento Labeledit
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Labeledit {
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
        Timestamp?: string;
        LabelID?: string;
        Action?: {
          name?: string;
          color?: number;
          predefinedID?: number;
          deleted?: boolean;
          orderIndex?: number;
          isActive?: boolean;
          type?: number;
          isImmutable?: boolean;
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
