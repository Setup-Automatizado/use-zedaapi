/**
 * Interface TypeScript para evento z_api/received_callback/reaction
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackReaction {
  headers?: {
    connection?: string;
    'content-length'?: string;
    host?: string;
    'http2-settings'?: string;
    upgrade?: string;
    'content-type'?: string;
    origin?: string;
    server?: string;
    'user-agent'?: string;
    'funnelchat-token'?: string;
  };
  body?: {
    isStatusReply?: boolean;
    chatLid?: string | null;
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
    senderPhoto?: string | null;
    senderName?: string;
    photo?: string | null;
    broadcast?: boolean;
    participantPhone?: string;
    participantLid?: string;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    reaction?: {
      value?: string;
      time?: number;
      reactionBy?: string;
      referencedMessage?: {
        messageId?: string;
        fromMe?: boolean;
        phone?: string;
        participant?: string;
      };
    };
  };
}
