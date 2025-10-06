import { z } from 'zod';

/**
 * Schema Zod para evento Deletechat
 * Gerado automaticamente pelo Webhook Mapper
 */
export const DeletechatSchema = z.object({
  headers: z.object({
    host: z.string(),
    'user-agent': z.string(),
    'content-length': z.string(),
    accept: z.string(),
    'content-type': z.string(),
    'accept-encoding': z.string(),
  }),
  body: z.object({
    data: z.object({
      event: z.object({
        JID: z.string(),
        Timestamp: z.string(),
        Action: z.object({
          messageRange: z.object({
            lastMessageTimestamp: z.number().optional(),
            messages: z
              .array(
                z.object({
                  key: z.object({
                    remoteJID: z.string(),
                    fromMe: z.boolean(),
                    ID: z.string(),
                    participant: z.string().optional(),
                  }),
                  timestamp: z.number(),
                })
              )
              .optional(),
            lastSystemMessageTimestamp: z.number().optional(),
          }),
        }),
        FromFullSync: z.boolean(),
      }),
      type: z.string(),
    }),
    eventType: z.string(),
    timestamp: z.number(),
    token: z.string(),
    userID: z.string(),
    userJID: z.string(),
    userName: z.string(),
    baseURL: z.string().optional(),
  }),
});

export type Deletechat = z.infer<typeof DeletechatSchema>;
