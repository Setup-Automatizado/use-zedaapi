/**
 * Interface TypeScript para evento z_api/received_callback/sticker
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackSticker {
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
    'z-api-token'?: string;
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
    photo?: string;
    broadcast?: boolean;
    participantLid?: string | null;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    sticker?: {
      stickerUrl?: string;
      mimeType?: string;
      width?: number;
      height?: number;
    };
    participantPhone?: string;
    referenceMessageId?: string;
    messageExpirationSeconds?: number;
  };
}
