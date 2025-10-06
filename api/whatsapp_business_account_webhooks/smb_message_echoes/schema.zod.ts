import { z } from 'zod';

/**
 * Schema Zod para evento whatsapp_business_account/smb_message_echoes
 * Gerado automaticamente pelo Webhook Mapper
 */
export const WhatsappBusinessAccountSmbMessageEchoesSchema = z.object({
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
              message_echoes: z.array(
                z.object({
                  from: z.string(),
                  to: z.string(),
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

export type WhatsappBusinessAccountSmbMessageEchoes = z.infer<
  typeof WhatsappBusinessAccountSmbMessageEchoesSchema
>;
