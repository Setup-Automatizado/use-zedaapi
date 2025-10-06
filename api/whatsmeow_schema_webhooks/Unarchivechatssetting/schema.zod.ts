import { z } from 'zod';

/**
 * Schema Zod para evento Unarchivechatssetting
 * Gerado automaticamente pelo Webhook Mapper
 */
export const UnarchivechatssettingSchema = z.object({
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
        Timestamp: z.string(),
        Action: z.object({
          unarchiveChats: z.boolean(),
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

export type Unarchivechatssetting = z.infer<typeof UnarchivechatssettingSchema>;
