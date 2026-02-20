// =============================================================================
// ZedaAPI Types - Matching exact Go API responses
// Based on: api/internal/http/handlers/partners.go + instances.go
// =============================================================================

// ---------------------------------------------------------------------------
// Partner API - POST /instances/integrator/on-demand (Bearer auth)
// ---------------------------------------------------------------------------

export interface CreateInstanceRequest {
	name: string;
	sessionName?: string;
	deliveryCallbackUrl?: string;
	receivedCallbackUrl?: string;
	isDevice?: boolean;
	businessDevice?: boolean;
	callRejectAuto?: boolean;
	autoReadMessage?: boolean;
}

export interface CreateInstanceResponse {
	id: string;
	token: string;
	due: string;
	instanceId: string;
	name: string;
	subscriptionActive: boolean;
	middleware: Record<string, unknown> | null;
	webhooks: Record<string, unknown> | null;
	createdAt: string;
}

// ---------------------------------------------------------------------------
// Partner API - GET /instances (Bearer auth)
// ---------------------------------------------------------------------------

export interface ZedaAPIInstance {
	id: string;
	token: string;
	name: string;
	created: string;
	phoneConnected: boolean;
	whatsappConnected: boolean;
	subscriptionActive: boolean;
	middleware: Record<string, unknown> | null;
	deliveryCallbackUrl: string;
	receivedCallbackUrl: string;
	receivedDeliveryUrl: string;
	messageStatusUrl: string;
	chatPresenceUrl: string;
	connectedUrl: string;
	disconnectedUrl: string;
}

export interface PaginatedResponse<T> {
	total: number;
	totalPage: number;
	pageSize: number;
	page: number;
	content: T[];
}

export interface ListInstancesParams {
	page?: number;
	pageSize?: number;
}

// ---------------------------------------------------------------------------
// Instance API - /instances/{id}/token/{token}/... (Client-Token auth)
// ---------------------------------------------------------------------------

/** GET /status */
export interface InstanceStatusResponse {
	connected: boolean;
	connectionStatus: string;
	smartphoneConnected: boolean;
	error: string | null;
}

/** GET /qr-code */
export interface QRCodeResponse {
	value: string;
}

/** GET /device */
export interface DeviceInfo {
	phone: string;
	platform: string;
	pushName: string;
	connected: boolean;
	waBrowser: string;
	waVersion: string;
}

/** GET /phone-code/{phone} */
export interface PhoneCodeResponse {
	code: string;
}

/** PUT /update-webhook-* individual endpoints */
export interface WebhookUpdateRequest {
	value: string;
}

/** PUT /update-every-webhooks */
export interface WebhookAllUpdateRequest {
	value: string;
	notifySentByMe?: boolean;
}

// ---------------------------------------------------------------------------
// Webhook event payloads (incoming from ZedaAPI callbacks)
// ---------------------------------------------------------------------------

export type ZedaAPIWebhookEventType =
	| "connected"
	| "disconnected"
	| "message"
	| "message_status"
	| "chat_presence"
	| "qr_code"
	| "ready"
	| "error";

export interface ZedaAPIWebhookEvent {
	instanceId: string;
	event: ZedaAPIWebhookEventType;
	data: Record<string, unknown>;
	timestamp: string;
}

// ---------------------------------------------------------------------------
// DLQ (Dead Letter Queue) types
// ---------------------------------------------------------------------------

export interface ZedaAPIDLQEntry {
	id: string;
	originalSubject: string;
	errorMessage: string;
	retryCount: number;
	payload: string;
	createdAt: string;
}

export interface ZedaAPIDLQListResponse {
	entries: ZedaAPIDLQEntry[];
	total: number;
}

// ---------------------------------------------------------------------------
// Error responses
// ---------------------------------------------------------------------------

export interface ZedaAPIErrorResponse {
	error: string;
	message?: string;
	statusCode?: number;
}
