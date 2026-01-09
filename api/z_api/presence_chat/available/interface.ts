/**
 * Interface TypeScript para evento z_api/presence_chat/available
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiPresenceChatAvailable {
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
    type?: string;
    phone?: string;
    status?: string;
    lastSeen?: null;
    instanceId?: string;
    participant?: string;
    participantLid?: string;
    participantPhone?: string;
    chatName?: string;
    photo?: string;
    senderName?: string;
    senderPhoto?: string;
    isGroup?: boolean;
  };
}
