/**
 * Interface TypeScript para evento GroupUpdate
 * Gerado automaticamente pelo Webhook Mapper
 */
export interface Groupupdate {
  type?: string;
  instanceId?: string;
  messageId?: string;
  phone?: string;
  sticker?: {
    stickerUrl?: string;
    mimeType?: string;
  };
  isGroup?: boolean;
  participantPhone?: string;
}
