import { z } from 'zod';

/**
 * Schema Zod para evento StatusUpdate
 * Gerado automaticamente pelo Webhook Mapper
 */
export const StatusupdateSchema = z.object({
  headers: z.object({
    accept: z.string(),
    'content-type': z.string(),
    'x-hub-signature': z.string(),
    'x-hub-signature-256': z.string(),
    'user-agent': z.string(),
    authorization: z.string(),
    'content-length': z.string(),
    'accept-encoding': z.string(),
    host: z.string(),
    connection: z.string(),
  }),
  body: z.object({
    object: z.string(),
    entry: z.array(
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
                  conversation: z
                    .object({
                      id: z.string(),
                      origin: z.object({
                        type: z.string(),
                      }),
                      expiration_timestamp: z.string().optional(),
                    })
                    .optional(),
                  pricing: z
                    .object({
                      billable: z.boolean(),
                      pricing_model: z.string(),
                      category: z.string(),
                      type: z.string(),
                    })
                    .optional(),
                })
              ),
            }),
            field: z.string(),
          })
        ),
      })
    ),
  }),
});

export type Statusupdate = z.infer<typeof StatusupdateSchema>;
