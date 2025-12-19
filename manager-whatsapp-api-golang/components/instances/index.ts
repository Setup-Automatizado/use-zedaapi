/**
 * Instance Components
 *
 * Re-exports all instance-related UI components for the WhatsApp dashboard.
 * Import from this barrel file for consistent component access.
 *
 * @example
 * ```typescript
 * import {
 *   InstanceTable,
 *   InstanceFilters,
 *   CreateInstanceButton,
 *   WebhookConfigForm,
 *   InstanceSettingsForm
 * } from '@/components/instances';
 * ```
 */

export { InstanceStatusBadge } from "./instance-status-badge";
export type { InstanceStatusBadgeProps } from "./instance-status-badge";

export { InstanceActionsDropdown } from "./instance-actions-dropdown";
export type { InstanceActionsDropdownProps } from "./instance-actions-dropdown";

export { InstanceFilters } from "./instance-filters";
export type { InstanceFiltersProps } from "./instance-filters";

export { InstanceTable } from "./instance-table";
export type { InstanceTableProps } from "./instance-table";

export { InstanceCard } from "./instance-card";
export type { InstanceCardProps } from "./instance-card";

export { CreateInstanceButton } from "./create-instance-button";
export type { CreateInstanceButtonProps } from "./create-instance-button";

// Webhook configuration components
export { WebhookConfigForm } from "./webhook-config-form";
export type { WebhookConfigFormProps } from "./webhook-config-form";

export { WebhookField } from "./webhook-field";
export type { WebhookFieldProps } from "./webhook-field";

// Settings components
export { InstanceSettingsForm } from "./instance-settings-form";
export type { InstanceSettingsFormProps } from "./instance-settings-form";

// Overview components
export { InstanceOverview } from "./instance-overview";
export type { InstanceOverviewProps } from "./instance-overview";
