import { z } from 'zod';

/**
 * Schema Zod para evento Pushname
 * Gerado automaticamente pelo Webhook Mapper
 */
export const PushnameSchema = z.object({
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
        Message: z.object({
          Chat: z.string(),
          Sender: z.string(),
          IsFromMe: z.boolean(),
          IsGroup: z.boolean(),
          AddressingMode: z.string(),
          SenderAlt: z.string(),
          RecipientAlt: z.string(),
          BroadcastListOwner: z.string(),
          ID: z.string(),
          ServerID: z.number(),
          Type: z.string(),
          PushName: z.string(),
          Timestamp: z.string(),
          Category: z.string(),
          Multicast: z.boolean(),
          MediaType: z.string(),
          Edit: z.string(),
          MsgBotInfo: z.object({
            EditType: z.string(),
            EditTargetID: z.string(),
            EditSenderTimestampMS: z.string(),
          }),
          MsgMetaInfo: z.object({
            TargetID: z.string(),
            TargetSender: z.string(),
            TargetChat: z.string(),
            DeprecatedLIDSession: z.null(),
            ThreadMessageID: z.string(),
            ThreadMessageSenderJID: z.string(),
          }),
          VerifiedName: z.union([z.object({}), z.null()]),
          DeviceSentMeta: z.null(),
        }),
        OldPushName: z.string(),
        NewPushName: z.string(),
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

export type Pushname = z.infer<typeof PushnameSchema>;
