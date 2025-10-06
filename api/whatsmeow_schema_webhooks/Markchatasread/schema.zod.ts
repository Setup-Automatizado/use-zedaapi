import { z } from 'zod';

/**
 * Schema Zod para evento Markchatasread
 * Gerado automaticamente pelo Webhook Mapper
 */
export const MarkchatasreadSchema = z.object({
  headers: z.object({
    host: z.string(),
    'user-agent': z.string(),
    'content-length': z.string(),
    accept: z.string(),
    'content-type': z.string(),
    'accept-encoding': z.string(),
  }),
  body: z.object({
    baseURL: z.string(),
    data: z.object({
      event: z.object({
        JID: z.string(),
        Timestamp: z.string(),
        Action: z.object({
          read: z.boolean(),
          messageRange: z.object({}),
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
  }),
});

export type Markchatasread = z.infer<typeof MarkchatasreadSchema>;
