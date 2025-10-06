import { z } from 'zod';

/**
 * Schema Zod para evento Keepalivetimeout
 * Gerado automaticamente pelo Webhook Mapper
 */
export const KeepalivetimeoutSchema = z.object({
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
        ErrorCount: z.number(),
        LastSuccess: z.string(),
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

export type Keepalivetimeout = z.infer<typeof KeepalivetimeoutSchema>;
