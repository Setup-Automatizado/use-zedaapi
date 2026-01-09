/**
 * Interface TypeScript para evento z_api/message_status/read
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiMessageStatusRead {
  headers?: {
    'content-length'?: string;
    host?: string;
    'content-type'?: string;
    origin?: string;
    server?: string;
    'user-agent'?: string;
    'funnelchat-token'?: string;
  };
  body?: {
    instanceId?: string;
    status?: string;
    ids?: string[];
    momment?: number;
    phoneDevice?: number;
    phone?: string;
    type?: string;
    isGroup?: boolean;
    participant?: string;
    participantDevice?: number;
  };
}
