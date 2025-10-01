import { z } from 'zod';

/**
 * Schema Zod para evento z_api/received_callback/document
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiReceivedCallbackDocumentSchema = z.object({
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
    'z-api-token': z.string(),
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
    senderPhoto: z.union([z.string()]),
    senderName: z.string(),
    photo: z.string(),
    broadcast: z.boolean(),
    participantLid: z.union([z.string()]),
    forwarded: z.boolean(),
    type: z.string(),
    fromApi: z.boolean(),
    document: z.object({
      caption: z.union([z.null(), z.string()]),
      documentUrl: z.string(),
      mimeType: z.string(),
      title: z.string(),
      pageCount: z.number(),
      fileName: z.string(),
    }),
    participantPhone: z.string().optional(),
    messageExpirationSeconds: z.number().optional(),
  }),
});

export type ZApiReceivedCallbackDocument = z.infer<typeof ZApiReceivedCallbackDocumentSchema>;
