import { z } from 'zod';

/**
 * Schema Zod para evento Appstate
 * Gerado automaticamente pelo Webhook Mapper
 */
export const AppstateSchema = z.object({
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
        Index: z.array(z.string()),
        timestamp: z.number(),
        callLogAction: z.object({
          callLogRecord: z.object({
            callResult: z.number(),
            isDndMode: z.boolean(),
            silenceReason: z.number(),
            duration: z.number(),
            startTime: z.number(),
            isIncoming: z.boolean(),
            isVideo: z.boolean(),
            callID: z.string(),
            callCreatorJID: z.string(),
            participants: z.array(
              z.object({
                userJID: z.string(),
                callResult: z.number(),
              })
            ),
            callType: z.number(),
          }),
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

export type Appstate = z.infer<typeof AppstateSchema>;
