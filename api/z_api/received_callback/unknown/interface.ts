/**
 * Interface TypeScript para evento z_api/received_callback/unknown
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface ZApiReceivedCallbackUnknown {
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
    isStatusReply?: boolean;
    chatLid?: string;
    connectedPhone?: string;
    waitingMessage?: boolean;
    isEdit?: boolean;
    isGroup?: boolean;
    isNewsletter?: boolean;
    instanceId?: string;
    messageId?: string;
    phone?: string;
    fromMe?: boolean;
    momment?: number;
    status?: string;
    chatName?: string;
    senderPhoto?: string | null;
    senderName?: string;
    photo?: string | null;
    broadcast?: boolean;
    participantLid?: null;
    forwarded?: boolean;
    type?: string;
    fromApi?: boolean;
    event?: {
      name?: string;
      description?: string;
      canceled?: boolean;
      joinLink?: string;
      scheduleTime?: number;
      location?: {
        name?: string;
      };
    };
    expiresAt?: null;
    referenceMessageId?: null | string;
    externalAdReply?: null;
    notification?: string;
    notificationParameters?: any[];
    callId?: string | null;
    code?: null;
    listMessage?: {
      description?: string;
      footerText?: string;
      title?: string;
      buttonText?: string;
      sections?: {
        title?: string;
        options?: {
          title?: string;
          description?: string;
          rowId?: string;
        }[];
      }[];
    };
    messageExpirationSeconds?: number;
    listResponseMessage?: {
      message?: string;
      title?: string;
      selectedRowId?: string;
    };
    buttonsMessage?: {
      imageUrl?: null;
      videoUrl?: null;
      message?: string;
      buttons?: {
        buttonId?: string;
        type?: number;
        buttonText?: {
          displayText?: string;
        };
      }[];
    };
    ptv?: {
      ptvUrl?: string;
      caption?: string;
      mimeType?: string;
      width?: number;
      height?: number;
    };
    liveLocation?: {
      longitude?: number;
      latitude?: number;
      sequence?: {
        low?: number;
        high?: number;
        unsigned?: boolean;
      };
      caption?: string;
    };
  };
}
