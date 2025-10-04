import { z } from 'zod';

/**
 * Schema Zod para evento Joinedgroup
 * Gerado automaticamente pelo Webhook Mapper
 */
export const JoinedgroupSchema = z.object({
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
        Reason: z.string(),
        Type: z.string(),
        CreateKey: z.string(),
        Sender: z.union([z.null(), z.string()]),
        SenderPN: z.null(),
        Notify: z.string(),
        JID: z.string(),
        OwnerJID: z.string(),
        OwnerPN: z.string(),
        Name: z.string(),
        NameSetAt: z.string(),
        NameSetBy: z.string(),
        NameSetByPN: z.string(),
        Topic: z.string(),
        TopicID: z.string(),
        TopicSetAt: z.string(),
        TopicSetBy: z.string(),
        TopicSetByPN: z.string(),
        TopicDeleted: z.boolean(),
        IsLocked: z.boolean(),
        IsAnnounce: z.boolean(),
        AnnounceVersionID: z.string(),
        IsEphemeral: z.boolean(),
        DisappearingTimer: z.number(),
        IsIncognito: z.boolean(),
        IsParent: z.boolean(),
        DefaultMembershipApprovalMode: z.string(),
        LinkedParentJID: z.string(),
        IsDefaultSubGroup: z.boolean(),
        IsJoinApprovalRequired: z.boolean(),
        AddressingMode: z.string(),
        GroupCreated: z.string(),
        CreatorCountryCode: z.string(),
        ParticipantVersionID: z.string(),
        Participants: z.array(
          z.object({
            JID: z.string(),
            PhoneNumber: z.string(),
            LID: z.string(),
            IsAdmin: z.boolean(),
            IsSuperAdmin: z.boolean(),
            DisplayName: z.string(),
            Error: z.number(),
            AddRequest: z.null(),
          })
        ),
        MemberAddMode: z.string(),
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

export type Joinedgroup = z.infer<typeof JoinedgroupSchema>;
