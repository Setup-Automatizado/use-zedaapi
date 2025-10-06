import { z } from 'zod';

/**
 * Schema Zod para evento generic/webhook
 * Gerado automaticamente pelo Webhook Mapper
 */
export const GenericWebhookSchema = z.object({
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
        changes: z
          .array(
            z.object({
              value: z.object({
                messaging_product: z.string(),
                metadata: z.object({
                  display_phone_number: z.string(),
                  phone_number_id: z.string(),
                }),
                message_echoes: z
                  .array(
                    z.object({
                      from: z.string(),
                      to: z.string(),
                      id: z.string(),
                      timestamp: z.string(),
                      type: z.string(),
                      document: z
                        .object({
                          filename: z.string(),
                          mime_type: z.string(),
                          sha256: z.string(),
                          id: z.string(),
                        })
                        .optional(),
                      video: z
                        .object({
                          mime_type: z.string(),
                          sha256: z.string(),
                          id: z.string(),
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
                      text: z
                        .object({
                          body: z.string(),
                        })
                        .optional(),
                      image: z
                        .object({
                          caption: z.string().optional(),
                          mime_type: z.string(),
                          sha256: z.string(),
                          id: z.string(),
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
                      button: z
                        .object({
                          text: z.string(),
                        })
                        .optional(),
                    })
                  )
                  .optional(),
                state_sync: z
                  .array(
                    z.object({
                      type: z.string(),
                      contact: z.object({
                        phone_number: z.string(),
                        full_name: z.string().optional(),
                        first_name: z.string().optional(),
                      }),
                      action: z.string(),
                      metadata: z.object({
                        timestamp: z.string(),
                      }),
                    })
                  )
                  .optional(),
              }),
              field: z.string(),
            })
          )
          .optional(),
        time: z.number().optional(),
        uid: z.string().optional(),
        changed_fields: z.array(z.string()).optional(),
      })
    ),
  }),
});

export type GenericWebhook = z.infer<typeof GenericWebhookSchema>;
