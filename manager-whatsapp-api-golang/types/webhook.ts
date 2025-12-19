/**
 * Webhook Configuration Types
 *
 * Type definitions for webhook configuration and management.
 * Supports various event types for WhatsApp message lifecycle.
 */

/**
 * Webhook event types
 * Each type corresponds to a specific WhatsApp event callback
 */
export type WebhookType =
  | 'delivery'              // Message delivery confirmations
  | 'received'              // Incoming messages
  | 'received-delivery'     // Combined received and delivery events
  | 'message-status'        // Message status updates (sent, delivered, read)
  | 'connected'             // Instance connected to WhatsApp
  | 'disconnected'          // Instance disconnected from WhatsApp
  | 'chat-presence';        // Chat participant presence updates (online/offline/typing)

/**
 * Webhook configuration settings
 * Complete set of webhook URLs and notification preferences
 */
export interface WebhookSettings {
  /** Delivery confirmation webhook URL */
  deliveryCallbackUrl?: string;

  /** Incoming message webhook URL */
  receivedCallbackUrl?: string;

  /** Combined received and delivery webhook URL */
  receivedAndDeliveryCallbackUrl?: string;

  /** Message status update webhook URL */
  messageStatusCallbackUrl?: string;

  /** Instance disconnected webhook URL */
  disconnectedCallbackUrl?: string;

  /** Chat presence update webhook URL */
  presenceChatCallbackUrl?: string;

  /** Instance connected webhook URL */
  connectedCallbackUrl?: string;

  /** Send webhooks for messages sent by the instance itself */
  notifySentByMe: boolean;
}

/**
 * Request to update a single webhook URL
 */
export interface WebhookUpdateRequest {
  /** New webhook URL (empty string to disable) */
  value: string;
}

/**
 * Response after updating a webhook URL
 */
export interface WebhookUpdateResponse {
  /** Update success status */
  value: boolean;

  /** Complete updated webhook configuration */
  webhooks: WebhookSettings;
}

/**
 * Request to update notify-sent-by-me setting
 */
export interface NotifySentByMeRequest {
  /** Enable/disable webhooks for own messages */
  notifySentByMe: boolean;
}

/**
 * Request to update all webhook URLs at once
 */
export interface AllWebhooksUpdateRequest {
  /** URL to apply to all webhook types */
  value: string;

  /** Optional notify-sent-by-me setting */
  notifySentByMe?: boolean;
}

/**
 * Webhook event payload structure (base)
 * Common fields across all webhook events
 */
export interface WebhookEventBase {
  /** Instance identifier */
  instanceId: string;

  /** Event timestamp (ISO 8601) */
  timestamp: string;

  /** Event type identifier */
  event: string;
}

/**
 * Webhook delivery configuration
 * Retry and timeout settings for webhook delivery
 */
export interface WebhookDeliveryConfig {
  /** Maximum retry attempts */
  maxRetries: number;

  /** Initial retry delay in milliseconds */
  retryDelayMs: number;

  /** Maximum retry delay in milliseconds */
  maxRetryDelayMs: number;

  /** Request timeout in milliseconds */
  timeoutMs: number;

  /** Exponential backoff multiplier */
  backoffMultiplier: number;
}

/**
 * Webhook validation result
 */
export interface WebhookValidation {
  /** Validation success status */
  valid: boolean;

  /** Validation error message if invalid */
  error?: string;

  /** Validated URL (normalized) */
  url?: string;
}

/**
 * Webhook test request
 */
export interface WebhookTestRequest {
  /** Webhook URL to test */
  url: string;

  /** Event type to simulate */
  eventType: WebhookType;

  /** Optional custom payload */
  payload?: Record<string, unknown>;
}

/**
 * Webhook test response
 */
export interface WebhookTestResponse {
  /** Test success status */
  success: boolean;

  /** HTTP status code received */
  statusCode?: number;

  /** Response body */
  responseBody?: string;

  /** Response time in milliseconds */
  responseTimeMs?: number;

  /** Error message if failed */
  error?: string;
}

/**
 * Mapping of webhook types to configuration fields
 */
export const WEBHOOK_TYPE_MAP: Record<WebhookType, keyof WebhookSettings> = {
  'delivery': 'deliveryCallbackUrl',
  'received': 'receivedCallbackUrl',
  'received-delivery': 'receivedAndDeliveryCallbackUrl',
  'message-status': 'messageStatusCallbackUrl',
  'connected': 'connectedCallbackUrl',
  'disconnected': 'disconnectedCallbackUrl',
  'chat-presence': 'presenceChatCallbackUrl',
} as const;

/**
 * Validate webhook URL format
 */
export function validateWebhookUrl(url: string): WebhookValidation {
  if (!url || url.trim() === '') {
    return { valid: true, url: '' }; // Empty is valid (disables webhook)
  }

  try {
    const parsed = new URL(url);

    if (!['http:', 'https:'].includes(parsed.protocol)) {
      return {
        valid: false,
        error: 'Webhook URL must use HTTP or HTTPS protocol',
      };
    }

    return { valid: true, url: parsed.toString() };
  } catch {
    return {
      valid: false,
      error: 'Invalid URL format',
    };
  }
}

/**
 * Check if webhook is configured
 */
export function isWebhookConfigured(
  settings: WebhookSettings,
  type: WebhookType
): boolean {
  const field = WEBHOOK_TYPE_MAP[type];
  const url = settings[field];
  return typeof url === 'string' && url.trim() !== '';
}

/**
 * Get all configured webhook types
 */
export function getConfiguredWebhooks(settings: WebhookSettings): WebhookType[] {
  return (Object.keys(WEBHOOK_TYPE_MAP) as WebhookType[]).filter(type =>
    isWebhookConfigured(settings, type)
  );
}

/**
 * Count configured webhooks
 */
export function countConfiguredWebhooks(settings: WebhookSettings): number {
  return getConfiguredWebhooks(settings).length;
}

/**
 * Check if any webhook is configured
 */
export function hasAnyWebhook(settings: WebhookSettings): boolean {
  return countConfiguredWebhooks(settings) > 0;
}
