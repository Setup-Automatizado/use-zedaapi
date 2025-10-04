/**
 * Interface TypeScript para evento Joinedgroup
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Joinedgroup {
  headers?: {
    host?: string;
    'user-agent'?: string;
    'content-length'?: string;
    accept?: string;
    'content-type'?: string;
    'accept-encoding'?: string;
  };
  body?: {
    baseURL?: string;
    data?: {
      event?: {
        Reason?: string;
        Type?: string;
        CreateKey?: string;
        Sender?: null | string;
        SenderPN?: null;
        Notify?: string;
        JID?: string;
        OwnerJID?: string;
        OwnerPN?: string;
        Name?: string;
        NameSetAt?: string;
        NameSetBy?: string;
        NameSetByPN?: string;
        Topic?: string;
        TopicID?: string;
        TopicSetAt?: string;
        TopicSetBy?: string;
        TopicSetByPN?: string;
        TopicDeleted?: boolean;
        IsLocked?: boolean;
        IsAnnounce?: boolean;
        AnnounceVersionID?: string;
        IsEphemeral?: boolean;
        DisappearingTimer?: number;
        IsIncognito?: boolean;
        IsParent?: boolean;
        DefaultMembershipApprovalMode?: string;
        LinkedParentJID?: string;
        IsDefaultSubGroup?: boolean;
        IsJoinApprovalRequired?: boolean;
        AddressingMode?: string;
        GroupCreated?: string;
        CreatorCountryCode?: string;
        ParticipantVersionID?: string;
        Participants?: {
          JID?: string;
          PhoneNumber?: string;
          LID?: string;
          IsAdmin?: boolean;
          IsSuperAdmin?: boolean;
          DisplayName?: string;
          Error?: number;
          AddRequest?: null;
        }[];
        MemberAddMode?: string;
      };
      type?: string;
    };
    eventType?: string;
    timestamp?: number;
    token?: string;
    userID?: string;
    userJID?: string;
    userName?: string;
  };
}
