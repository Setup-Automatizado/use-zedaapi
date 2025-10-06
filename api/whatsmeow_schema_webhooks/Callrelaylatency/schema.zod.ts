import { z } from 'zod';

/**
 * Schema Zod para evento Callrelaylatency
 * Gerado automaticamente pelo Webhook Mapper
 */
export const CallrelaylatencySchema = z.object({
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
        From: z.string(),
        Timestamp: z.string(),
        CallCreator: z.string(),
        CallID: z.string(),
        GroupJID: z.string(),
        Data: z.object({
          Tag: z.string(),
          Attrs: z.object({
            'call-creator': z.string(),
            'call-id': z.string(),
          }),
          Content: z.array(
            z.object({
              Tag: z.string(),
              Attrs: z.object({
                latency: z.string().optional(),
                dl_bw: z.string().optional(),
                xlatency: z.string().optional(),
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
    baseURL: z.string().optional(),
  }),
});

export type Callrelaylatency = z.infer<typeof CallrelaylatencySchema>;
