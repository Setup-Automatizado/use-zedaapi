/**
 * Interface TypeScript para evento StatusUpdate
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Statusupdate {
  headers?: {
    accept?: string;
    'content-type'?: string;
    'x-hub-signature'?: string;
    'x-hub-signature-256'?: string;
    'user-agent'?: string;
    authorization?: string;
    'content-length'?: string;
    'accept-encoding'?: string;
    host?: string;
    connection?: string;
  };
  body?: {
    object?: string;
    entry?: {
      id?: string;
      changes?: {
        value?: {
          messaging_product?: string;
          metadata?: {
            display_phone_number?: string;
            phone_number_id?: string;
          };
          statuses?: {
            id?: string;
            status?: string;
            timestamp?: string;
            recipient_id?: string;
            conversation?: {
              id?: string;
              origin?: {
                type?: string;
              };
              expiration_timestamp?: string;
            };
            pricing?: {
              billable?: boolean;
              pricing_model?: string;
              category?: string;
              type?: string;
            };
          }[];
        };
        field?: string;
      }[];
    }[];
  };
}
