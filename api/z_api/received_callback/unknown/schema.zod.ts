import { z } from 'zod';

/**
 * Schema Zod para evento z_api/received_callback/unknown
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiReceivedCallbackUnknownSchema = z.object({
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
    isStatusReply: z.boolean().optional(),
    chatLid: z.union([z.string()]),
    connectedPhone: z.string(),
    waitingMessage: z.boolean().optional(),
    isEdit: z.boolean().optional(),
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
    participantLid: z.null().optional(),
    forwarded: z.boolean(),
    type: z.string(),
    fromApi: z.boolean().optional(),
    event: z
      .object({
        name: z.string(),
        description: z.string(),
        canceled: z.boolean(),
        joinLink: z.string(),
        scheduleTime: z.number(),
        location: z.object({
          name: z.string(),
        }),
      })
      .optional(),
    expiresAt: z.null().optional(),
    referenceMessageId: z.union([z.null(), z.string()]).optional(),
    externalAdReply: z.null().optional(),
    notification: z.string().optional(),
    notificationParameters: z.array(z.any()).optional(),
    callId: z.union([z.string(), z.null()]).optional(),
    code: z.null().optional(),
    listMessage: z
      .object({
        description: z.string(),
        footerText: z.string(),
        title: z.string(),
        buttonText: z.string(),
        sections: z.array(
          z.object({
            title: z.string(),
            options: z.array(
              z.object({
                title: z.string(),
                description: z.string(),
                rowId: z.string(),
              })
            ),
          })
        ),
      })
      .optional(),
    messageExpirationSeconds: z.number().optional(),
    listResponseMessage: z
      .object({
        message: z.string(),
        title: z.string(),
        selectedRowId: z.string(),
      })
      .optional(),
    buttonsMessage: z
      .object({
        imageUrl: z.null(),
        videoUrl: z.null(),
        message: z.string(),
        buttons: z.array(
          z.object({
            buttonId: z.string(),
            type: z.number(),
            buttonText: z.object({
              displayText: z.string(),
            }),
          })
        ),
      })
      .optional(),
    ptv: z
      .object({
        ptvUrl: z.string(),
        caption: z.string(),
        mimeType: z.string(),
        width: z.number(),
        height: z.number(),
      })
      .optional(),
    liveLocation: z
      .object({
        longitude: z.number(),
        latitude: z.number(),
        sequence: z.object({
          low: z.number(),
          high: z.number(),
          unsigned: z.boolean(),
        }),
        caption: z.string(),
      })
      .optional(),
  }),
});

export type ZApiReceivedCallbackUnknown = z.infer<typeof ZApiReceivedCallbackUnknownSchema>;
