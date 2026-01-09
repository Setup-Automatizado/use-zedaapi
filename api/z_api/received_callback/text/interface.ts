/**
 * Interface TypeScript para evento z_api/received_callback/text
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackText {
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
    chatLid?: null | string;
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
    participantLid?: string | null;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    text?: {
      message?: string;
      description?: string;
      title?: string;
      url?: string;
      /** TRUNCATED FIELD - Original: text */
      thumbnailUrl?: string;
    };
    messageExpirationSeconds?: number;
    referenceMessageId?: string;
    editMessageId?: string;
  };
}
