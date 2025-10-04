import { z } from 'zod';

/**
 * Schema Zod para evento Archive
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ArchiveSchema = z.object({
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
        Timestamp: z.string(),
        Action: z.object({
          archived: z.boolean(),
          messageRange: z.object({
            lastMessageTimestamp: z.number(),
            messages: z.array(
              z.object({
                key: z.object({
                  remoteJID: z.string(),
                  fromMe: z.boolean(),
                  ID: z.string(),
                }),
                timestamp: z.number(),
              })
            ),
          }),
        }),
        FromFullSync: z.boolean(),
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

export type Archive = z.infer<typeof ArchiveSchema>;
