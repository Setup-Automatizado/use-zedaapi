/**
 * Interface TypeScript para evento z_api/received_callback/poll
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackPoll {
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
    chatLid?: null;
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
    participantPhone?: string;
    participantLid?: string;
    messageExpirationSeconds?: number;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    poll?: {
      question?: string;
      pollMaxOptions?: number;
      options?: {
        name?: string;
      }[];
    };
  };
}
