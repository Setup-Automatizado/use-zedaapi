/**
 * Custom Hooks Barrel Export
 *
 * Centralized export for all custom React hooks used in the WhatsApp Instance Manager.
 * Import hooks from this barrel file for consistent access across the application.
 *
 * @example
 * ```typescript
 * import { useInstances, useInstanceStatus, useQRCode } from '@/hooks';
 * ```
 */

// Generic hooks
export { usePolling } from "./use-polling";
export type { UsePollingOptions, UsePollingResult } from "./use-polling";

// Instance hooks
export { useInstances } from "./use-instances";
export type { UseInstancesParams, UseInstancesResult } from "./use-instances";

export { useInstance } from "./use-instance";
export type { UseInstanceResult } from "./use-instance";

export { useInstanceStatus } from "./use-instance-status";
export type {
	UseInstanceStatusOptions,
	UseInstanceStatusResult,
} from "./use-instance-status";

// QR code hook
export { useQRCode } from "./use-qr-code";
export type { UseQRCodeOptions, UseQRCodeResult } from "./use-qr-code";

// Health check hook
export { useHealth } from "./use-health";
export type { UseHealthOptions, UseHealthResult } from "./use-health";

// Media query hook
export { useMediaQuery } from "./use-media-query";
