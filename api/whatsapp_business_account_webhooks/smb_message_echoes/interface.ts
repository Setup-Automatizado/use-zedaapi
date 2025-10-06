/**
 * Interface TypeScript para evento whatsapp_business_account/smb_message_echoes
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface WhatsappBusinessAccountSmbMessageEchoes {
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
