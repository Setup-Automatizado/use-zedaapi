import { z } from 'zod';

/**
 * Schema Zod para evento Groupinfo
 * Gerado automaticamente pelo Webhook Mapper
 */
export const GroupinfoSchema = z.object({
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
        JID: z.string(),
        Notify: z.string(),
        Sender: z.union([z.string()]),
        SenderPN: z.union([z.null(), z.string()]),
        Timestamp: z.string(),
        Name: z.null(),
        Topic: z.union([z.object({}), z.null()]),
        Locked: z.null(),
        Announce: z.union([z.null(), z.object({})]),
        Ephemeral: z.null(),
        MembershipApprovalMode: z.null(),
        Delete: z.null(),
        Link: z.null(),
        Unlink: z.union([z.null(), z.object({})]),
        NewInviteLink: z.union([z.string(), z.null()]),
        PrevParticipantVersionID: z.string(),
        ParticipantVersionID: z.string(),
        JoinReason: z.string(),
        Join: z.union([z.array(z.any()), z.null()]),
        Leave: z.union([z.array(z.any())]),
        Promote: z.union([z.null(), z.array(z.any())]),
        Demote: z.null(),
        UnknownChanges: z.null(),
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

export type Groupinfo = z.infer<typeof GroupinfoSchema>;
