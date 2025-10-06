/**
 * Interface TypeScript para evento Appstate
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Appstate {
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
        Index?: string[];
        timestamp?: number;
        callLogAction?: {
          callLogRecord?: {
            callResult?: number;
            isDndMode?: boolean;
            silenceReason?: number;
            duration?: number;
            startTime?: number;
            isIncoming?: boolean;
            isVideo?: boolean;
            callID?: string;
            callCreatorJID?: string;
            participants?: {
              userJID?: string;
              callResult?: number;
            }[];
            callType?: number;
          };
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
