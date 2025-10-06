/**
 * Interface TypeScript para evento Callterminate
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Callterminate {
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
        Reason?: string;
        Data?: {
          Tag?: string;
          Attrs?: {
            'call-creator'?: string;
            'call-id'?: string;
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
    baseURL?: string;
  };
}
