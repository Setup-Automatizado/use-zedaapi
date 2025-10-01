import { z } from 'zod';

/**
 * Schema Zod para evento z_api/received_callback/poll
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiReceivedCallbackPollSchema = z.object({
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
    senderPhoto: z.string(),
    senderName: z.string(),
    photo: z.string(),
    broadcast: z.boolean(),
    participantPhone: z.string(),
    participantLid: z.string(),
    messageExpirationSeconds: z.number(),
    forwarded: z.boolean(),
    type: z.string(),
    fromApi: z.boolean(),
    poll: z.object({
      question: z.string(),
      pollMaxOptions: z.number(),
      options: z.array(
        z.object({
          name: z.string(),
        })
      ),
    }),
  }),
});

export type ZApiReceivedCallbackPoll = z.infer<typeof ZApiReceivedCallbackPollSchema>;
