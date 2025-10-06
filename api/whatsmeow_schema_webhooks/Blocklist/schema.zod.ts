import { z } from 'zod';

/**
 * Schema Zod para evento Blocklist
 * Gerado automaticamente pelo Webhook Mapper
 */
export const BlocklistSchema = z.object({
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
        Action: z.string(),
        DHash: z.string(),
        PrevDHash: z.string(),
        Changes: z.array(
          z.object({
            JID: z.string(),
            Action: z.string(),
          })
        ),
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

export type Blocklist = z.infer<typeof BlocklistSchema>;
