/**
 * Interface TypeScript para evento z_api/received_callback/video
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackVideo {
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
    participantLid?: string;
    messageExpirationSeconds?: number;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    video?: {
      videoUrl?: string;
      caption?: string;
      mimeType?: string;
      width?: number;
      height?: number;
      seconds?: number;
      viewOnce?: boolean;
      isGif?: boolean;
    };
    editMessageId?: string;
    referenceMessageId?: string;
  };
}
