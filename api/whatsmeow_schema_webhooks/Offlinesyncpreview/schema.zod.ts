import { z } from 'zod';

/**
 * Schema Zod para evento Offlinesyncpreview
 * Gerado automaticamente pelo Webhook Mapper
 */
export const OfflinesyncpreviewSchema = z.object({
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
        Total: z.number(),
        AppDataChanges: z.number(),
        Messages: z.number(),
        Notifications: z.number(),
        Receipts: z.number(),
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

export type Offlinesyncpreview = z.infer<typeof OfflinesyncpreviewSchema>;
