import { z } from 'zod';

/**
 * Schema Zod para evento whatsapp_business_account/messages_unsupported
 * Gerado automaticamente pelo Webhook Mapper
 */
export const WhatsappBusinessAccountMessagesUnsupportedSchema = z.object({
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
              contacts: z.array(
                z.object({
                  profile: z.object({
                    name: z.string(),
                  }),
                  wa_id: z.string(),
                })
              ),
              messages: z.array(
                z.object({
                  from: z.string(),
                  id: z.string(),
                  timestamp: z.string(),
                  errors: z.array(
                    z.object({
                      code: z.number(),
                      title: z.string(),
                      message: z.string(),
                      error_data: z.object({
                        details: z.string(),
                      }),
                    })
                  ),
                  type: z.string(),
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

export type WhatsappBusinessAccountMessagesUnsupported = z.infer<
  typeof WhatsappBusinessAccountMessagesUnsupportedSchema
>;
