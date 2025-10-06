import { z } from 'zod';

/**
 * Schema Zod para evento Businessname
 * Gerado automaticamente pelo Webhook Mapper
 */
export const BusinessnameSchema = z.object({
  data: z
    .object({
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
          VerifiedName: z.object({
            Certificate: z.object({
              details: z.string(),
              signature: z.string(),
            }),
            Details: z.object({
              serial: z.number(),
              issuer: z.string(),
              verifiedName: z.string(),
            }),
          }),
          DeviceSentMeta: z.null(),
        }),
        OldBusinessName: z.string(),
        NewBusinessName: z.string(),
      }),
      type: z.string(),
    })
    .optional(),
  eventType: z.string().optional(),
  timestamp: z.number().optional(),
  token: z.string().optional(),
  userID: z.string().optional(),
  userJID: z.string().optional(),
  userName: z.string().optional(),
  headers: z
    .object({
      host: z.string(),
      'user-agent': z.string(),
      'content-length': z.string(),
      accept: z.string(),
      'content-type': z.string(),
      'accept-encoding': z.string(),
    })
    .optional(),
  body: z
    .object({
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
            VerifiedName: z.object({
              Certificate: z.object({
                details: z.string(),
                signature: z.string(),
                serverSignature: z.string().optional(),
              }),
              Details: z.object({
                serial: z.number(),
                issuer: z.string(),
                verifiedName: z.string(),
                issueTime: z.number().optional(),
              }),
            }),
            DeviceSentMeta: z.null(),
          }),
          OldBusinessName: z.string(),
          NewBusinessName: z.string(),
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
    })
    .optional(),
});

export type Businessname = z.infer<typeof BusinessnameSchema>;
