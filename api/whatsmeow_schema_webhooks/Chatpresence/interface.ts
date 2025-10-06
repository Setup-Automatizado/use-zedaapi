/**
 * Interface TypeScript para evento Chatpresence
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Chatpresence {
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
        Chat?: string;
        Sender?: string;
        IsFromMe?: boolean;
        IsGroup?: boolean;
        AddressingMode?: string;
        SenderAlt?: string;
        RecipientAlt?: string;
        BroadcastListOwner?: string;
        State?: string;
        Media?: string;
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
