import { z } from 'zod';

/**
 * Schema Zod para evento Calltransport
 * Gerado automaticamente pelo Webhook Mapper
 */
export const CalltransportSchema = z.object({
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
        From: z.string(),
        Timestamp: z.string(),
        CallCreator: z.string(),
        CallID: z.string(),
        GroupJID: z.string(),
        RemotePlatform: z.string(),
        RemoteVersion: z.string(),
        Data: z.object({
          Tag: z.string(),
          Attrs: z.object({
            'call-creator': z.string(),
            'call-id': z.string(),
            'transport-message-type': z.string(),
          }),
          Content: z.array(
            z.object({
              Tag: z.string(),
              Attrs: z.object({
                priority: z.string(),
              }),
              Content: z.string(),
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

export type Calltransport = z.infer<typeof CalltransportSchema>;
