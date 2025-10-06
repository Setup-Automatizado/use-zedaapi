import { z } from 'zod';

/**
* Schema Zod para evento Qr
* Gerado automaticamente pelo Webhook Mapper
*/
export const QrSchema = z.object({
  headers: z.object({
    host: z.string(),
    "user-agent": z.string(),
    "content-length": z.string(),
    accept: z.string(),
    "content-type": z.string(),
    "accept-encoding": z.string()
  }),
  body: z.object({
    data: z.object({
      code: z.string(),
      event: z.string(),
      /** TRUNCATED FIELD - Original: text */,
      qrCodeBase64: z.string().describe('TRUNCATED FIELD - Original type: text'),
      type: z.string()
    }),
    eventType: z.string(),
    timestamp: z.number(),
    token: z.string(),
    userID: z.string(),
    userJID: z.string(),
    userName: z.string()
  })
});

export type Qr = z.infer<typeof QrSchema>;
