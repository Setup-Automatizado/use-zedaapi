import { z } from 'zod';

/**
 * Schema Zod para evento z_api/received_callback/video
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiReceivedCallbackVideoSchema = z.object({
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
    chatLid: z.union([z.null(), z.string()]),
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
    photo: z.union([z.string(), z.null()]),
    broadcast: z.boolean(),
    participantPhone: z.string().optional(),
    participantLid: z.union([z.string()]),
    messageExpirationSeconds: z.number().optional(),
    forwarded: z.boolean(),
    type: z.string(),
    fromApi: z.boolean(),
    video: z.object({
      videoUrl: z.union([z.string()]),
      caption: z.string(),
      mimeType: z.string(),
      width: z.number(),
      height: z.number(),
      seconds: z.number(),
      viewOnce: z.boolean(),
      isGif: z.boolean(),
    }),
    editMessageId: z.string().optional(),
    referenceMessageId: z.string().optional(),
  }),
});

export type ZApiReceivedCallbackVideo = z.infer<typeof ZApiReceivedCallbackVideoSchema>;
