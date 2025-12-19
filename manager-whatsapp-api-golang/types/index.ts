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

// Instance types
export type {
	Instance,
	InstanceData,
	InstanceMiddleware,
	InstanceStatus,
	DeviceInfo,
	DeviceDetails,
	InstanceListResponse,
	RawInstanceListResponse,
	CreateInstanceRequest,
	CreateInstanceResponse,
	QRCodeResponse,
	PhonePairingResponse,
	InstanceSettings,
	WebhookSettings as InstanceWebhookSettings,
} from "./instance";

export {
	normalizeInstance,
	normalizeInstances,
	isInstanceConnected,
	hasActiveSubscription,
	getInstanceId,
	getInstanceToken,
} from "./instance";

// Health check types
export type {
	HealthStatus,
	CircuitState,
	HealthResponse,
	ReadinessResponse,
	HealthChecks,
	ComponentStatus,
	DetailedHealthResponse,
	HealthCheckOptions,
} from "./health";

export {
	isHealthy,
	isDegraded,
	isComponentOperational,
	getCriticalStatus,
	getAverageHealthCheckDuration,
} from "./health";

// Webhook types
export type {
	WebhookType,
	WebhookSettings,
	WebhookUpdateRequest,
	WebhookUpdateResponse,
	NotifySentByMeRequest,
	AllWebhooksUpdateRequest,
	WebhookEventBase,
	WebhookDeliveryConfig,
	WebhookValidation,
	WebhookTestRequest,
	WebhookTestResponse,
} from "./webhook";

export {
	WEBHOOK_TYPE_MAP,
	validateWebhookUrl,
	isWebhookConfigured,
	getConfiguredWebhooks,
	countConfiguredWebhooks,
	hasAnyWebhook,
} from "./webhook";

// Generic API types
export type {
	ActionResult,
	PaginationParams,
	PaginatedResponse,
	ApiError,
	ApiResponse,
	ResponseMetadata,
	SortOptions,
	FilterOptions,
	DateRange,
	BatchRequest,
	BatchResponse,
	BatchResult,
	HttpMethod,
	RequestConfig,
	ApiClientConfig,
} from "./api";

export {
	isSuccess,
	isError,
	success,
	error,
	validationError,
	unwrap,
	unwrapOr,
	calculatePagination,
	validatePaginationParams,
} from "./api";
