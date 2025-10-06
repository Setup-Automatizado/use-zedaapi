import { z } from 'zod';

/**
 * Schema Zod para evento GroupUpdate
 * Gerado automaticamente pelo Webhook Mapper
 */
export const GroupupdateSchema = z.object({
  type: z.string().optional(),
  instanceId: z.string().optional(),
  messageId: z.string().optional(),
  phone: z.string().optional(),
  sticker: z
    .object({
      stickerUrl: z.string(),
      mimeType: z.string(),
    })
    .optional(),
  isGroup: z.boolean().optional(),
  participantPhone: z.string().optional(),
});

export type Groupupdate = z.infer<typeof GroupupdateSchema>;
