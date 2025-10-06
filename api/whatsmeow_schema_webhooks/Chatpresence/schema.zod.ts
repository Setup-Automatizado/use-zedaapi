import { z } from 'zod';

/**
 * Schema Zod para evento Chatpresence
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ChatpresenceSchema = z.object({
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
        Chat: z.string(),
        Sender: z.string(),
        IsFromMe: z.boolean(),
        IsGroup: z.boolean(),
        AddressingMode: z.string(),
        SenderAlt: z.string(),
        RecipientAlt: z.string(),
        BroadcastListOwner: z.string(),
        State: z.string(),
        Media: z.string(),
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

export type Chatpresence = z.infer<typeof ChatpresenceSchema>;
