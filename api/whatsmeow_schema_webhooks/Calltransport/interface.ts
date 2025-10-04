/**
 * Interface TypeScript para evento Calltransport
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Calltransport {
  headers?: {
    host?: string;
    'user-agent'?: string;
    'content-length'?: string;
    accept?: string;
    'content-type'?: string;
    'accept-encoding'?: string;
  };
  body?: {
    baseURL?: string;
    data?: {
      event?: {
        From?: string;
        Timestamp?: string;
        CallCreator?: string;
        CallID?: string;
        GroupJID?: string;
        RemotePlatform?: string;
        RemoteVersion?: string;
        Data?: {
          Tag?: string;
          Attrs?: {
            'call-creator'?: string;
            'call-id'?: string;
            'transport-message-type'?: string;
          };
          Content?: {
            Tag?: string;
            Attrs?: {
              priority?: string;
            };
            Content?: string;
          }[];
        };
      };
      type?: string;
    };
    eventType?: string;
    timestamp?: number;
    token?: string;
    userID?: string;
    userJID?: string;
    userName?: string;
  };
}
