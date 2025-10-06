import { z } from 'zod';

/**
 * Schema Zod para evento ProfileUpdate
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ProfileupdateSchema = z.object({
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
              messages: z.array(
                z.object({
                  from: z.string(),
                  id: z.string(),
                  timestamp: z.string(),
                  text: z
                    .object({
                      body: z.string(),
                    })
                    .optional(),
                  type: z.string(),
                  image: z
                    .object({
                      mime_type: z.string(),
                      sha256: z.string(),
                      id: z.string(),
                      caption: z.string().optional(),
                    })
                    .optional(),
                  audio: z
                    .object({
                      mime_type: z.string(),
                      sha256: z.string(),
                      id: z.string(),
                      voice: z.boolean(),
                    })
                    .optional(),
                  context: z
                    .object({
                      from: z.string(),
                      id: z.string().optional(),
                    })
                    .optional(),
                  sticker: z
                    .object({
                      mime_type: z.string(),
                      sha256: z.string(),
                      id: z.string(),
                      animated: z.boolean(),
                    })
                    .optional(),
                  reaction: z
                    .object({
                      message_id: z.string(),
                      emoji: z.string(),
                    })
                    .optional(),
                  errors: z
                    .array(
                      z.object({
                        code: z.number(),
                        title: z.string(),
                        message: z.string(),
                        error_data: z.object({
                          details: z.string(),
                        }),
                      })
                    )
                    .optional(),
                  interactive: z
                    .object({
                      type: z.string(),
                      list_reply: z.object({
                        id: z.string(),
                        title: z.string(),
                        description: z.string(),
                      }),
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

export type Profileupdate = z.infer<typeof ProfileupdateSchema>;
