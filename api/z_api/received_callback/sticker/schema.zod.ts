import { z } from 'zod';

/**
 * Schema Zod para evento z_api/received_callback/sticker
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiReceivedCallbackStickerSchema = z.object({
  headers: z.object({
    connection: z.string(),
    'content-length': z.string(),
    host: z.string(),
    'http2-settings': z.string(),
    upgrade: z.string(),
    'content-type': z.string(),
    origin: z.string(),
    server: z.string(),
    'user-agent': z.string(),
    'funnelchat-token': z.string(),
  }),
  body: z.object({
    isStatusReply: z.boolean(),
    chatLid: z.union([z.string(), z.null()]),
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
    photo: z.union([z.string()]),
    broadcast: z.boolean(),
    participantLid: z.union([z.string(), z.null()]),
    forwarded: z.boolean(),
    type: z.string(),
    fromApi: z.boolean(),
    sticker: z.object({
      stickerUrl: z.string(),
      mimeType: z.string(),
      width: z.number(),
      height: z.number(),
    }),
    participantPhone: z.string().optional(),
    referenceMessageId: z.string().optional(),
    messageExpirationSeconds: z.number().optional(),
  }),
});

export type ZApiReceivedCallbackSticker = z.infer<typeof ZApiReceivedCallbackStickerSchema>;
