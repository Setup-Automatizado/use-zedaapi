import { z } from 'zod';

/**
 * Schema Zod para evento Appstatesynccomplete
 * Gerado automaticamente pelo Webhook Mapper
 */
export const AppstatesynccompleteSchema = z.object({
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
        Name: z.string(),
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

export type Appstatesynccomplete = z.infer<typeof AppstatesynccompleteSchema>;
