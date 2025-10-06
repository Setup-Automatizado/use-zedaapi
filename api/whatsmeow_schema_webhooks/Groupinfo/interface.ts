/**
 * Interface TypeScript para evento Groupinfo
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Groupinfo {
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
        JID?: string;
        Notify?: string;
        Sender?: string;
        SenderPN?: null | string;
        Timestamp?: string;
        Name?: null;
        Topic?: object | null;
        Locked?: null;
        Announce?: null | object;
        Ephemeral?: null;
        MembershipApprovalMode?: null;
        Delete?: null;
        Link?: null;
        Unlink?: null | object;
        NewInviteLink?: string | null;
        PrevParticipantVersionID?: string;
        ParticipantVersionID?: string;
        JoinReason?: string;
        Join?: any[] | null;
        Leave?: any[];
        Promote?: null | any[];
        Demote?: null;
        UnknownChanges?: null;
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
