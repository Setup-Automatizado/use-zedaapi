/**
 * Interface TypeScript para evento ProfileUpdate
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Profileupdate {
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
          contacts?: {
            profile?: {
              name?: string;
            };
            wa_id?: string;
          }[];
          messages?: {
            from?: string;
            id?: string;
            timestamp?: string;
            text?: {
              body?: string;
            };
            type?: string;
            image?: {
              mime_type?: string;
              sha256?: string;
              id?: string;
              caption?: string;
            };
            audio?: {
              mime_type?: string;
              sha256?: string;
              id?: string;
              voice?: boolean;
            };
            context?: {
              from?: string;
              id?: string;
            };
            sticker?: {
              mime_type?: string;
              sha256?: string;
              id?: string;
              animated?: boolean;
            };
            reaction?: {
              message_id?: string;
              emoji?: string;
            };
            errors?: {
              code?: number;
              title?: string;
              message?: string;
              error_data?: {
                details?: string;
              };
            }[];
            interactive?: {
              type?: string;
              list_reply?: {
                id?: string;
                title?: string;
                description?: string;
              };
            };
          }[];
        };
        field?: string;
      }[];
    }[];
  };
}
