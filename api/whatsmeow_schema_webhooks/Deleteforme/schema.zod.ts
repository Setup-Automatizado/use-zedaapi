import { z } from 'zod';

/**
 * Schema Zod para evento Deleteforme
 * Gerado automaticamente pelo Webhook Mapper
 */
export const DeleteformeSchema = z.object({
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
        ChatJID: z.string(),
        SenderJID: z.string(),
        IsFromMe: z.boolean(),
        MessageID: z.string(),
        Timestamp: z.string(),
        Action: z.object({
          deleteMedia: z.boolean(),
          messageTimestamp: z.number(),
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

export type Deleteforme = z.infer<typeof DeleteformeSchema>;
