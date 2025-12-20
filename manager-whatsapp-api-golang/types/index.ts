/**
 * WhatsApp API Type Definitions
 *
 * Centralized export for all type definitions used in the WhatsApp Instance Manager.
 * Import from this barrel file for consistent type access across the application.
 *
 * @example
 * ```typescript
 * import { Instance, WebhookSettings, ActionResult } from '@/types';
 * ```
 */

// Generic API types
export type {
	ActionResult,
	ApiClientConfig,
	ApiError,
	ApiResponse,
	BatchRequest,
	BatchResponse,
	BatchResult,
	DateRange,
	FilterOptions,
	HttpMethod,
	PaginatedResponse,
	PaginationParams,
	RequestConfig,
	ResponseMetadata,
	SortOptions,
} from "./api";
export {
	calculatePagination,
	error,
	isError,
	isSuccess,
	success,
	unwrap,
	unwrapOr,
	validatePaginationParams,
	validationError,
} from "./api";

// Health check types
export type {
	CircuitState,
	ComponentStatus,
	DetailedHealthResponse,
	HealthCheckOptions,
	HealthChecks,
	HealthResponse,
	HealthStatus,
	ReadinessResponse,
} from "./health";

export {
	getAverageHealthCheckDuration,
	getCriticalStatus,
	isComponentOperational,
	isDegraded,
	isHealthy,
} from "./health";
// Instance types
export type {
	CreateInstanceRequest,
	CreateInstanceResponse,
	DeviceDetails,
	DeviceInfo,
	Instance,
	InstanceData,
	InstanceListResponse,
	InstanceMiddleware,
	InstanceSettings,
	InstanceStatus,
	PhonePairingResponse,
	QRCodeResponse,
	RawInstanceListResponse,
	WebhookSettings as InstanceWebhookSettings,
} from "./instance";
export {
	getInstanceId,
	getInstanceToken,
	hasActiveSubscription,
	isInstanceConnected,
	normalizeInstance,
	normalizeInstances,
} from "./instance";
// Webhook types
export type {
	AllWebhooksUpdateRequest,
	NotifySentByMeRequest,
	WebhookDeliveryConfig,
	WebhookEventBase,
	WebhookSettings,
	WebhookTestRequest,
	WebhookTestResponse,
	WebhookType,
	WebhookUpdateRequest,
	WebhookUpdateResponse,
	WebhookValidation,
} from "./webhook";
export {
	countConfiguredWebhooks,
	getConfiguredWebhooks,
	hasAnyWebhook,
	isWebhookConfigured,
	validateWebhookUrl,
	WEBHOOK_TYPE_MAP,
} from "./webhook";
