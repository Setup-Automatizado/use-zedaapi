import { z } from 'zod';

/**
 * Schema Zod para evento Mute
 * Gerado automaticamente pelo Webhook Mapper
 */
export const MuteSchema = z.object({
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
          muted: z.boolean(),
          muteEndTimestamp: z.number().optional(),
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

export type Mute = z.infer<typeof MuteSchema>;
