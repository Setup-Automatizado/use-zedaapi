import { z } from 'zod';

/**
 * Schema Zod para evento CallEvent
 * Gerado automaticamente pelo Webhook Mapper
 */
export const CalleventSchema = z.object({
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
              contacts: z.array(
                z.object({
                  profile: z.object({
                    name: z.string(),
                  }),
                  wa_id: z.string(),
                })
              ),
              calls: z
                .array(
                  z.object({
                    id: z.string(),
                    from: z.string(),
                    to: z.string(),
                    event: z.string(),
                    timestamp: z.string(),
                    direction: z.string(),
                    session: z
                      .object({
                        sdp: z.string(),
                        sdp_type: z.string(),
                      })
                      .optional(),
                    status: z.string().optional(),
                  })
                )
                .optional(),
              messages: z
                .array(
                  z.object({
                    context: z.object({
                      from: z.string(),
                      id: z.string(),
                    }),
                    from: z.string(),
                    id: z.string(),
                    timestamp: z.string(),
                    type: z.string(),
                    interactive: z.object({
                      type: z.string(),
                      call_permission_reply: z.object({
                        response: z.string(),
                        expiration_timestamp: z.number(),
                        response_source: z.string(),
                      }),
                    }),
                  })
                )
                .optional(),
            }),
            field: z.string(),
          })
        ),
      })
    ),
  }),
});

export type Callevent = z.infer<typeof CalleventSchema>;
