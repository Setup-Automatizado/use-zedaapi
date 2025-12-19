export const APP_NAME = 'WhatsApp Manager';
export const DEFAULT_PAGE_SIZE = 20;
export const QR_CODE_REFRESH_INTERVAL = 30000; // 30s
export const STATUS_POLL_INTERVAL = 5000; // 5s
export const HEALTH_POLL_INTERVAL = 30000; // 30s

export const INSTANCE_STATUS = {
  CONNECTED: 'connected',
  DISCONNECTED: 'disconnected',
  PENDING: 'pending',
  ERROR: 'error',
} as const;

export type InstanceStatus = typeof INSTANCE_STATUS[keyof typeof INSTANCE_STATUS];

export const WEBHOOK_TYPES = [
  {
    key: 'delivery',
    label: 'Delivery',
    description: 'When message is delivered'
  },
  {
    key: 'received',
    label: 'Received',
    description: 'When message is received'
  },
  {
    key: 'received-delivery',
    label: 'Received & Delivery',
    description: 'Received and delivered'
  },
  {
    key: 'message-status',
    label: 'Message Status',
    description: 'Message status change'
  },
  {
    key: 'connected',
    label: 'Connected',
    description: 'When instance connects'
  },
  {
    key: 'disconnected',
    label: 'Disconnected',
    description: 'When instance disconnects'
  },
  {
    key: 'chat-presence',
    label: 'Chat Presence',
    description: 'Chat presence status'
  },
] as const;

export type WebhookType = typeof WEBHOOK_TYPES[number]['key'];

export const STATUS_COLORS = {
  connected: {
    bg: 'bg-green-50 dark:bg-green-950/20',
    border: 'border-green-200 dark:border-green-800',
    text: 'text-green-700 dark:text-green-400',
    dot: 'bg-green-500',
  },
  disconnected: {
    bg: 'bg-red-50 dark:bg-red-950/20',
    border: 'border-red-200 dark:border-red-800',
    text: 'text-red-700 dark:text-red-400',
    dot: 'bg-red-500',
  },
  pending: {
    bg: 'bg-yellow-50 dark:bg-yellow-950/20',
    border: 'border-yellow-200 dark:border-yellow-800',
    text: 'text-yellow-700 dark:text-yellow-400',
    dot: 'bg-yellow-500',
  },
  error: {
    bg: 'bg-red-50 dark:bg-red-950/20',
    border: 'border-red-200 dark:border-red-800',
    text: 'text-red-700 dark:text-red-400',
    dot: 'bg-red-500',
  },
} as const;
