import { z } from 'zod';

/**
 * Schema Zod para evento Labeledit
 * Gerado automaticamente pelo Webhook Mapper
 */
export const LabeleditSchema = z.object({
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
        Timestamp: z.string(),
        LabelID: z.string(),
        Action: z.object({
          name: z.string(),
          color: z.number(),
          predefinedID: z.number(),
          deleted: z.boolean(),
          orderIndex: z.number(),
          isActive: z.boolean(),
          type: z.number(),
          isImmutable: z.boolean(),
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
    baseURL: z.string().optional(),
  }),
});

export type Labeledit = z.infer<typeof LabeleditSchema>;
