import { z } from 'zod';

/**
 * Schema Zod para evento z_api/message_status/read_by_me
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiMessageStatusReadByMeSchema = z.object({
  headers: z.object({
    'content-length': z.string(),
    host: z.string(),
    'content-type': z.string(),
    origin: z.string(),
    server: z.string(),
    'user-agent': z.string(),
    'z-api-token': z.string(),
  }),
  body: z.object({
    instanceId: z.string(),
    status: z.string(),
    ids: z.array(z.string()),
    momment: z.number(),
    phoneDevice: z.number(),
    phone: z.string(),
    type: z.string(),
    isGroup: z.boolean(),
    participant: z.string().optional(),
    participantDevice: z.number().optional(),
  }),
});

export type ZApiMessageStatusReadByMe = z.infer<typeof ZApiMessageStatusReadByMeSchema>;
