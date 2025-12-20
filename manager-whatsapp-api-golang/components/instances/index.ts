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

export type { CreateInstanceButtonProps } from "./create-instance-button";
export { CreateInstanceButton } from "./create-instance-button";
export type { InstanceActionsDropdownProps } from "./instance-actions-dropdown";
export { InstanceActionsDropdown } from "./instance-actions-dropdown";
export type { InstanceCardProps } from "./instance-card";
export { InstanceCard } from "./instance-card";
export type { InstanceFiltersProps } from "./instance-filters";
export { InstanceFilters } from "./instance-filters";
export type { InstanceOverviewProps } from "./instance-overview";
// Overview components
export { InstanceOverview } from "./instance-overview";
export type { InstanceSettingsFormProps } from "./instance-settings-form";
// Settings components
export { InstanceSettingsForm } from "./instance-settings-form";
export type { InstanceStatusBadgeProps } from "./instance-status-badge";
export { InstanceStatusBadge } from "./instance-status-badge";
export type { InstanceTableProps } from "./instance-table";
export { InstanceTable } from "./instance-table";
export type { WebhookConfigFormProps } from "./webhook-config-form";
// Webhook configuration components
export { WebhookConfigForm } from "./webhook-config-form";
export type { WebhookFieldProps } from "./webhook-field";
export { WebhookField } from "./webhook-field";
