import { z } from 'zod';

/**
 * Schema Zod para evento Connected
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ConnectedSchema = z.object({
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
        Timestamp: z.string().optional(),
        Action: z
          .object({
            name: z.string(),
          })
          .optional(),
        FromFullSync: z.boolean().optional(),
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

export type Connected = z.infer<typeof ConnectedSchema>;
