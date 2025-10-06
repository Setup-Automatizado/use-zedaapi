/**
 * Interface TypeScript para evento CallEvent
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Callevent {
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
          calls?: {
            id?: string;
            from?: string;
            to?: string;
            event?: string;
            timestamp?: string;
            direction?: string;
            session?: {
              sdp?: string;
              sdp_type?: string;
            };
            status?: string;
          }[];
          messages?: {
            context?: {
              from?: string;
              id?: string;
            };
            from?: string;
            id?: string;
            timestamp?: string;
            type?: string;
            interactive?: {
              type?: string;
              call_permission_reply?: {
                response?: string;
                expiration_timestamp?: number;
                response_source?: string;
              };
            };
          }[];
        };
        field?: string;
      }[];
    }[];
  };
}
