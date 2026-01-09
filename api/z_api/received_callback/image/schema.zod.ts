import { z } from 'zod';

/**
* Schema Zod para evento z_api/received_callback/image
* Gerado automaticamente pelo Webhook Mapper
*/
export const ZApiReceivedCallbackImageSchema = z.object({
  headers: z.object({
    connection: z.string(),
    "content-length": z.string(),
    host: z.string(),
    "http2-settings": z.string(),
    upgrade: z.string(),
    "content-type": z.string(),
    origin: z.string(),
    server: z.string(),
    "user-agent": z.string(),
    "funnelchat-token": z.string()
  }),
  body: z.object({
    isStatusReply: z.boolean(),
    chatLid: z.null(),
    connectedPhone: z.string(),
    waitingMessage: z.boolean(),
    isEdit: z.boolean(),
    isGroup: z.boolean(),
    isNewsletter: z.boolean(),
    instanceId: z.string(),
    messageId: z.string(),
    phone: z.string(),
    fromMe: z.boolean(),
    momment: z.number(),
    status: z.string(),
    chatName: z.string(),
    senderPhoto: z.union([z.string(), z.null()]),
    senderName: z.string(),
    photo: z.string(),
    broadcast: z.boolean(),
    participantPhone: z.string(),
    participantLid: z.string(),
    messageExpirationSeconds: z.number().optional(),
    forwarded: z.boolean(),
    type: z.string(),
    fromApi: z.boolean(),
    image: z.object({
      /** TRUNCATED FIELD - Original: text */,
      imageUrl: z.union([z.string(), z.null()]),
      /** TRUNCATED FIELD - Original: text */,
      thumbnailUrl: z.union([z.string(), z.null()]),
      caption: z.string(),
      mimeType: z.string(),
      viewOnce: z.boolean(),
      width: z.number(),
      height: z.number()
    }),
    referenceMessageId: z.string().optional(),
    editMessageId: z.string().optional()
  })
});

export type ZApiReceivedCallbackImage = z.infer<typeof ZApiReceivedCallbackImageSchema>;
