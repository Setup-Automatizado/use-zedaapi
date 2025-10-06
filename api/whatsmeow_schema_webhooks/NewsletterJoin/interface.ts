/**
 * Interface TypeScript para evento NewsletterJoin
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Newsletterjoin {
  isStatusReply?: boolean;
  chatLid?: string;
  connectedPhone?: string;
  waitingMessage?: boolean;
  isEdit?: boolean;
  isGroup?: boolean;
  isNewsletter?: boolean;
  instanceId?: string;
  messageId?: string;
  phone?: string;
  fromMe?: boolean;
  momment?: number;
  status?: string;
  chatName?: string;
  senderPhoto?: null;
  senderName?: string;
  photo?: string;
  broadcast?: boolean;
  participantLid?: null;
  forwarded?: boolean;
  type?: string;
  fromApi?: boolean;
  event?: {
    name: string;
    description: string;
    canceled: boolean;
    joinLink: string;
    scheduleTime: number;
    location: {};
  };
}
