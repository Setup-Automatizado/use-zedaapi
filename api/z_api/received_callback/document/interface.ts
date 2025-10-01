/**
 * Interface TypeScript para evento z_api/received_callback/document
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackDocument {
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
    senderPhoto?: string;
    senderName?: string;
    photo?: string;
    broadcast?: boolean;
    participantLid?: string;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    document?: {
      caption?: null | string;
      documentUrl?: string;
      mimeType?: string;
      title?: string;
      pageCount?: number;
      fileName?: string;
    };
    participantPhone?: string;
    messageExpirationSeconds?: number;
  };
}
