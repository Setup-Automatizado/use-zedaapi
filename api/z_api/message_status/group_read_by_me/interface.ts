/**
 * Interface TypeScript para evento z_api/message_status/group_read_by_me
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiMessageStatusGroupReadByMe {
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
    participant?: string;
    participantDevice?: number;
    phone?: string;
    type?: string;
    isGroup?: boolean;
  };
}
