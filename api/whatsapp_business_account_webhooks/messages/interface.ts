/**
 * Interface TypeScript para evento whatsapp_business_account/messages
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface WhatsappBusinessAccountMessages {
  object?: string;
  entry?: {
    id: string;
    changes: {
      value: {
        messaging_product: string;
        metadata: {
          display_phone_number: string;
          phone_number_id: string;
        };
        statuses: {
          id: string;
          status: string;
          timestamp: string;
          recipient_id: string;
        }[];
      };
      field: string;
    }[];
  }[];
}
