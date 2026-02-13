/**
 * Dashboard Components
 *
 * Barrel export for all dashboard-related components.
 * These components provide overview statistics, quick actions,
 * and health monitoring for the WhatsApp Instance Manager.
 */

export { HealthSummary, type HealthSummaryProps } from "./health-summary";
export { QuickActions } from "./quick-actions";
export {
	type DeviceMap,
	RecentInstances,
	type RecentInstancesProps,
} from "./recent-instances";
export { StatsCards, type StatsCardsProps } from "./stats-cards";
