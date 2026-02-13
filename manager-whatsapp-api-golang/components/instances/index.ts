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
export type { InstanceStatisticsProps } from "./instance-statistics";
export { InstanceStatistics } from "./instance-statistics";
export type { InstanceStatusBadgeProps } from "./instance-status-badge";
export { InstanceStatusBadge } from "./instance-status-badge";
export type { InstanceTableProps } from "./instance-table";
export { InstanceTable } from "./instance-table";
export type { MessageTestFormProps } from "./message-test-form";
export { MessageTestForm } from "./message-test-form";
export type { QueueStatusCardProps } from "./queue-status-card";
export { QueueStatusCard } from "./queue-status-card";
export type { SubscriptionManagementProps } from "./subscription-management";
export { SubscriptionManagement } from "./subscription-management";
export type { TokenDisplayProps } from "./token-display";
export { TokenDisplay } from "./token-display";
export type { ProxyConfigFormProps } from "./proxy-config-form";
// Proxy configuration components
export { ProxyConfigForm } from "./proxy-config-form";
export type { WebhookConfigFormProps } from "./webhook-config-form";
// Webhook configuration components
export { WebhookConfigForm } from "./webhook-config-form";
export type { WebhookFieldProps } from "./webhook-field";
export { WebhookField } from "./webhook-field";
export type { PoolProxyAssignmentProps } from "./pool-proxy-assignment";
export { PoolProxyAssignment } from "./pool-proxy-assignment";
