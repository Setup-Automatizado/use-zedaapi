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

export type { UseHealthOptions, UseHealthResult } from "./use-health";
// Health check hook
export { useHealth } from "./use-health";
export type { UseInstanceResult } from "./use-instance";
export { useInstance } from "./use-instance";
export type {
	UseInstanceStatusOptions,
	UseInstanceStatusResult,
} from "./use-instance-status";
export { useInstanceStatus } from "./use-instance-status";
export type { UseInstancesParams, UseInstancesResult } from "./use-instances";
// Instance hooks
export { useInstances } from "./use-instances";
// Media query hook
export { useMediaQuery } from "./use-media-query";
export type { UsePollingOptions, UsePollingResult } from "./use-polling";
// Generic hooks
export { usePolling } from "./use-polling";
export type { UseQRCodeOptions, UseQRCodeResult } from "./use-qr-code";
// QR code hook
export { useQRCode } from "./use-qr-code";
