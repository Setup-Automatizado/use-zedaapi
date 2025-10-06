import { z } from 'zod';

/**
 * Schema Zod para evento Keepaliverestored
 * Gerado automaticamente pelo Webhook Mapper
 */
export const KeepaliverestoredSchema = z.object({
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
      event: z.object({}),
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

export type Keepaliverestored = z.infer<typeof KeepaliverestoredSchema>;
