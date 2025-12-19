/**
 * Dashboard Components
 *
 * Barrel export for all dashboard-related components.
 * These components provide overview statistics, quick actions,
 * and health monitoring for the WhatsApp Instance Manager.
 */

export { StatsCards, type StatsCardsProps } from "./stats-cards";
export {
	RecentInstances,
	type RecentInstancesProps,
	type DeviceMap,
} from "./recent-instances";
export { QuickActions } from "./quick-actions";
export { HealthSummary, type HealthSummaryProps } from "./health-summary";
