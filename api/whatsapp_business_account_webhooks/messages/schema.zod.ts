import { z } from 'zod';

/**
 * Schema Zod para evento whatsapp_business_account/messages
 * Gerado automaticamente pelo Webhook Mapper
 */
export const WhatsappBusinessAccountMessagesSchema = z.object({
  object: z.string().optional(),
  entry: z
    .array(
      z.object({
        id: z.string(),
        changes: z.array(
          z.object({
            value: z.object({
              messaging_product: z.string(),
              metadata: z.object({
                display_phone_number: z.string(),
                phone_number_id: z.string(),
              }),
              statuses: z.array(
                z.object({
                  id: z.string(),
                  status: z.string(),
                  timestamp: z.string(),
                  recipient_id: z.string(),
                })
              ),
            }),
            field: z.string(),
          })
        ),
      })
    )
    .optional(),
});

export type WhatsappBusinessAccountMessages = z.infer<typeof WhatsappBusinessAccountMessagesSchema>;
