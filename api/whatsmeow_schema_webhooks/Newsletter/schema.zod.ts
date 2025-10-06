import { z } from 'zod';

/**
 * Schema Zod para evento Newsletter
 * Gerado automaticamente pelo Webhook Mapper
 */
export const NewsletterSchema = z.object({
  chatLid: z.null().optional(),
  isGroup: z.boolean().optional(),
  isNewsletter: z.boolean().optional(),
  instanceId: z.string().optional(),
  messageId: z.string().optional(),
  phone: z.string().optional(),
  connectedPhone: z.string().optional(),
  fromMe: z.boolean().optional(),
  momment: z.number().optional(),
  expiresAt: z.null().optional(),
  status: z.string().optional(),
  chatName: z.string().optional(),
  senderPhoto: z.null().optional(),
  senderName: z.string().optional(),
  photo: z.null().optional(),
  broadcast: z.boolean().optional(),
  referenceMessageId: z.null().optional(),
  externalAdReply: z.null().optional(),
  forwarded: z.boolean().optional(),
  type: z.string().optional(),
  notification: z.string().optional(),
  notificationParameters: z.array(z.string()).optional(),
  callId: z.null().optional(),
  code: z.null().optional(),
});

export type Newsletter = z.infer<typeof NewsletterSchema>;
