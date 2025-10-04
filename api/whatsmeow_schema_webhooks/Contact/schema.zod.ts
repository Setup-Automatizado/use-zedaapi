import { z } from 'zod';

/**
 * Schema Zod para evento Contact
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ContactSchema = z.object({
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
          fullName: z.string(),
          firstName: z.string().optional(),
          lidJID: z.string(),
          saveOnPrimaryAddressbook: z.boolean(),
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

export type Contact = z.infer<typeof ContactSchema>;
