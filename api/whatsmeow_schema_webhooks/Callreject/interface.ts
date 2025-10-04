/**
 * Interface TypeScript para evento Callreject
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Callreject {
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
        Data?: {
          Tag?: string;
          Attrs?: {
            'call-creator'?: string;
            'call-id'?: string;
            count?: string;
          };
          Content?: null;
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
