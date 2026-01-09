/**
 * Interface TypeScript para evento z_api/message_status/group_played
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiMessageStatusGroupPlayed {
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
    participant?: string;
    participantDevice?: number;
    phone?: string;
    type?: string;
    isGroup?: boolean;
  };
}
