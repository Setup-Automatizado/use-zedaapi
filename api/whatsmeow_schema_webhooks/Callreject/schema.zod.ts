import { z } from 'zod';

/**
 * Schema Zod para evento Callreject
 * Gerado automaticamente pelo Webhook Mapper
 */
export const CallrejectSchema = z.object({
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
        From: z.string(),
        Timestamp: z.string(),
        CallCreator: z.string(),
        CallID: z.string(),
        GroupJID: z.string(),
        Data: z.object({
          Tag: z.string(),
          Attrs: z.object({
            'call-creator': z.string(),
            'call-id': z.string(),
            count: z.string(),
          }),
          Content: z.null(),
        }),
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

export type Callreject = z.infer<typeof CallrejectSchema>;
