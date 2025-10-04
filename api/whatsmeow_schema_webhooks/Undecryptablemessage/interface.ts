/**
 * Interface TypeScript para evento Undecryptablemessage
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Undecryptablemessage {
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
        Info?: {
          Chat?: string;
          Sender?: string;
          IsFromMe?: boolean;
          IsGroup?: boolean;
          AddressingMode?: string;
          SenderAlt?: string;
          RecipientAlt?: string;
          BroadcastListOwner?: string;
          ID?: string;
          ServerID?: number;
          Type?: string;
          PushName?: string;
          Timestamp?: string;
          Category?: string;
          Multicast?: boolean;
          MediaType?: string;
          Edit?: string;
          MsgBotInfo?: {
            EditType?: string;
            EditTargetID?: string;
            EditSenderTimestampMS?: string;
          };
          MsgMetaInfo?: {
            TargetID?: string;
            TargetSender?: string;
            TargetChat?: string;
            DeprecatedLIDSession?: null;
            ThreadMessageID?: string;
            ThreadMessageSenderJID?: string;
          };
          VerifiedName?: object | null;
          DeviceSentMeta?: null;
        };
        IsUnavailable?: boolean;
        UnavailableType?: string;
        DecryptFailMode?: string;
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
