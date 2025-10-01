/**
 * Interface TypeScript para evento z_api/presence_chat/composing
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiPresenceChatComposing {
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
    type?: string;
    phone?: string;
    status?: string;
    lastSeen?: null;
    instanceId?: string;
  };
}
