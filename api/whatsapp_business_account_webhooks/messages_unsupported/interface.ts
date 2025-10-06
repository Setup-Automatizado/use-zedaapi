/**
 * Interface TypeScript para evento whatsapp_business_account/messages_unsupported
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface WhatsappBusinessAccountMessagesUnsupported {
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
          errors?: {
            code?: number;
            title?: string;
            message?: string;
            error_data?: {
              details?: string;
            };
          }[];
          type?: string;
        }[];
      };
      field?: string;
    }[];
  }[];
}
