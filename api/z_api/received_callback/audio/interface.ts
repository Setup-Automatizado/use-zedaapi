/**
 * Interface TypeScript para evento z_api/received_callback/audio
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackAudio {
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
    photo?: string | null;
    broadcast?: boolean;
    participantLid?: null | string;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    audio?: {
      ptt?: boolean;
      seconds?: number;
      audioUrl?: string;
      mimeType?: string;
      viewOnce?: boolean;
    };
    participantPhone?: string;
    referenceMessageId?: string;
    messageExpirationSeconds?: number;
  };
}
