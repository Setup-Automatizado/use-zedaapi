/**
 * Interface TypeScript para evento generic/webhook
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface GenericWebhook {
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
          message_echoes?: {
            from?: string;
            to?: string;
            id?: string;
            timestamp?: string;
            type?: string;
            document?: {
              filename?: string;
              mime_type?: string;
              sha256?: string;
              id?: string;
            };
            video?: {
              mime_type?: string;
              sha256?: string;
              id?: string;
            };
            audio?: {
              mime_type?: string;
              sha256?: string;
              id?: string;
              voice?: boolean;
            };
            text?: {
              body?: string;
            };
            image?: {
              caption?: string;
              mime_type?: string;
              sha256?: string;
              id?: string;
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
            };
            sticker?: {
              mime_type?: string;
              sha256?: string;
              id?: string;
              animated?: boolean;
            };
            button?: {
              text?: string;
            };
          }[];
          state_sync?: {
            type?: string;
            contact?: {
              phone_number?: string;
              full_name?: string;
              first_name?: string;
            };
            action?: string;
            metadata?: {
              timestamp?: string;
            };
          }[];
        };
        field?: string;
      }[];
      time?: number;
      uid?: string;
      changed_fields?: string[];
    }[];
  };
}
