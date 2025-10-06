import { z } from 'zod';

/**
 * Schema Zod para evento Userabout
 * Gerado automaticamente pelo Webhook Mapper
 */
export const UseraboutSchema = z.object({
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
        Status: z.string(),
        Timestamp: z.string(),
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

export type Userabout = z.infer<typeof UseraboutSchema>;
