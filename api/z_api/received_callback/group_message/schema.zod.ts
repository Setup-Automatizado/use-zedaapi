import { z } from 'zod';

/**
 * Schema Zod para evento z_api/received_callback/group_message
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiReceivedCallbackGroupMessageSchema = z.object({
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
    connectedPhone: z.string(),
    isGroup: z.boolean(),
    isNewsletter: z.boolean(),
    instanceId: z.string(),
    messageId: z.string(),
    momment: z.number(),
    status: z.string(),
    fromMe: z.boolean(),
    phone: z.string(),
    chatName: z.string(),
    senderName: z.string(),
    senderPhoto: z.union([z.string(), z.null()]),
    photo: z.union([z.string(), z.null()]),
    broadcast: z.boolean(),
    participantPhone: z.string().optional(),
    participantLid: z.null().optional(),
    type: z.string(),
    waitingMessage: z.boolean().optional(),
    viewOnce: z.boolean().optional(),
    chatLid: z.null().optional(),
    expiresAt: z.null().optional(),
    referenceMessageId: z.null().optional(),
    externalAdReply: z.null().optional(),
    forwarded: z.boolean().optional(),
    notification: z.string().optional(),
    notificationParameters: z.array(z.string()).optional(),
    callId: z.null().optional(),
    code: z.null().optional(),
  }),
});

export type ZApiReceivedCallbackGroupMessage = z.infer<
  typeof ZApiReceivedCallbackGroupMessageSchema
>;
