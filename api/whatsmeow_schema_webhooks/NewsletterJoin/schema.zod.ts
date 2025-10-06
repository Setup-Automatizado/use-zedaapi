import { z } from 'zod';

/**
 * Schema Zod para evento NewsletterJoin
 * Gerado automaticamente pelo Webhook Mapper
 */
export const NewsletterjoinSchema = z.object({
  isStatusReply: z.boolean().optional(),
  chatLid: z.string().optional(),
  connectedPhone: z.string().optional(),
  waitingMessage: z.boolean().optional(),
  isEdit: z.boolean().optional(),
  isGroup: z.boolean().optional(),
  isNewsletter: z.boolean().optional(),
  instanceId: z.string().optional(),
  messageId: z.string().optional(),
  phone: z.string().optional(),
  fromMe: z.boolean().optional(),
  momment: z.number().optional(),
  status: z.string().optional(),
  chatName: z.string().optional(),
  senderPhoto: z.null().optional(),
  senderName: z.string().optional(),
  photo: z.string().optional(),
  broadcast: z.boolean().optional(),
  participantLid: z.null().optional(),
  forwarded: z.boolean().optional(),
  type: z.string().optional(),
  fromApi: z.boolean().optional(),
  event: z
    .object({
      name: z.string(),
      description: z.string(),
      canceled: z.boolean(),
      joinLink: z.string(),
      scheduleTime: z.number(),
      location: z.object({}),
    })
    .optional(),
});

export type Newsletterjoin = z.infer<typeof NewsletterjoinSchema>;
