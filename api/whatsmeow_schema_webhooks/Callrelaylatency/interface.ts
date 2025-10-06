/**
 * Interface TypeScript para evento Callrelaylatency
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Callrelaylatency {
  headers?: {
    host?: string;
    'user-agent'?: string;
    'content-length'?: string;
    accept?: string;
    'content-type'?: string;
    'accept-encoding'?: string;
  };
  body?: {
    data?: {
      event?: {
        From?: string;
        Timestamp?: string;
        CallCreator?: string;
        CallID?: string;
        GroupJID?: string;
        Data?: {
          Tag?: string;
          Attrs?: {
            'call-creator'?: string;
            'call-id'?: string;
          };
          Content?: {
            Tag?: string;
            Attrs?: {
              latency?: string;
              dl_bw?: string;
              xlatency?: string;
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
    baseURL?: string;
  };
}
