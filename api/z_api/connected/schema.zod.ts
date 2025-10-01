import { z } from 'zod';

/**
 * Schema Zod para evento z_api/connected
 * Gerado automaticamente pelo Webhook Mapper
 */
export const ZApiConnectedSchema = z.object({
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
    connected: z.boolean(),
    momment: z.number(),
    instanceId: z.string(),
    phone: z.string(),
    isBusiness: z.boolean(),
  }),
});

export type ZApiConnected = z.infer<typeof ZApiConnectedSchema>;
