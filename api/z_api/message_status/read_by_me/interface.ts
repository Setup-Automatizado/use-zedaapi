/**
 * Interface TypeScript para evento z_api/message_status/read_by_me
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiMessageStatusReadByMe {
  headers?: {
    'content-length'?: string;
    host?: string;
    'content-type'?: string;
    origin?: string;
    server?: string;
    'user-agent'?: string;
    'z-api-token'?: string;
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
