/**
 * Interface TypeScript para evento z_api/received_callback/group_message
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackGroupMessage {
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
    connectedPhone?: string;
    isGroup?: boolean;
    isNewsletter?: boolean;
    instanceId?: string;
    messageId?: string;
    momment?: number;
    status?: string;
    fromMe?: boolean;
    phone?: string;
    chatName?: string;
    senderName?: string;
    senderPhoto?: string | null;
    photo?: string | null;
    broadcast?: boolean;
    participantPhone?: string;
    participantLid?: null;
    type?: string;
    waitingMessage?: boolean;
    viewOnce?: boolean;
    chatLid?: null;
    expiresAt?: null;
    referenceMessageId?: null;
    externalAdReply?: null;
    forwarded?: boolean;
    notification?: string;
    notificationParameters?: string[];
    callId?: null;
    code?: null;
  };
}
