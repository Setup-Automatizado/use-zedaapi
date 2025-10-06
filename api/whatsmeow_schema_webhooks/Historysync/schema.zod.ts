import { z } from 'zod';

/**
 * Schema Zod para evento Historysync
 * Gerado automaticamente pelo Webhook Mapper
 */
export const HistorysyncSchema = z.object({
  headers: z.object({
    host: z.string(),
    'user-agent': z.string(),
    'content-length': z.string(),
    accept: z.string(),
    'content-type': z.string(),
    'accept-encoding': z.string(),
  }),
  body: z.object({
    data: z.object({
      event: z.object({
        Data: z.object({
          syncType: z.number(),
          conversations: z.array(
            z.object({
              ID: z.string(),
              messages: z.array(
                z.object({
                  message: z.object({
                    key: z.object({
                      remoteJID: z.string(),
                      fromMe: z.boolean(),
                      ID: z.string(),
                    }),
                    message: z.object({
                      conversation: z.string(),
                      messageContextInfo: z.object({
                        messageSecret: z.string(),
                      }),
                    }),
                    messageTimestamp: z.number(),
                    participant: z.string(),
                    messageSecret: z.string(),
                    isMentionedInStatus: z.boolean(),
                  }),
                  msgOrderID: z.number(),
                })
              ),
              unreadCount: z.number(),
              readOnly: z.boolean(),
              ephemeralExpiration: z.number(),
              ephemeralSettingTimestamp: z.number(),
              endOfHistoryTransferType: z.number(),
              conversationTimestamp: z.number(),
              name: z.string(),
              pHash: z.string(),
              notSpam: z.boolean(),
              archived: z.boolean(),
              unreadMentionCount: z.number(),
              markedAsUnread: z.boolean(),
              suspended: z.boolean(),
            })
          ),
          chunkOrder: z.number(),
          progress: z.number(),
          phoneNumberToLidMappings: z.array(
            z.object({
              pnJID: z.string(),
              lidJID: z.string(),
            })
          ),
        }),
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

export type Historysync = z.infer<typeof HistorysyncSchema>;
