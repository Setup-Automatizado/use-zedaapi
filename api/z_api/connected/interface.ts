/**
 * Interface TypeScript para evento z_api/connected
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiConnected {
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
    connected?: boolean;
    momment?: number;
    instanceId?: string;
    phone?: string;
    isBusiness?: boolean;
  };
}
