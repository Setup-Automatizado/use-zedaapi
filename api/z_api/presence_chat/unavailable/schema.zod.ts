import { z } from 'zod';

/**
 * Schema Zod para evento z_api/presence_chat/unavailable
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiPresenceChatUnavailableSchema = z.object({
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
    type: z.string(),
    phone: z.string(),
    status: z.string(),
    lastSeen: z.null(),
    instanceId: z.string(),
  }),
});

export type ZApiPresenceChatUnavailable = z.infer<typeof ZApiPresenceChatUnavailableSchema>;
