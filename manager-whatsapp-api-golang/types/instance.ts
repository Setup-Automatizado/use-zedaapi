/**
 * WhatsApp Instance Types
 *
 * Type definitions for WhatsApp instance management, including
 * instance creation, configuration, status tracking, and device information.
 *
 * IMPORTANT: Field names must match the Go API JSON tags exactly:
 * - instanceId (not id)
 * - instanceToken (not token)
 * - createdAt (not created)
 *
 * We provide normalized aliases (id, token, created) for backward compatibility.
 */

/**
 * Middleware type for WhatsApp connection
 * @property web - Browser-based connection
 * @property mobile - Mobile device connection
 */
export type InstanceMiddleware = "web" | "mobile";

/**
 * Webhook settings from API
 */
export interface WebhookSettings {
	deliveryCallbackUrl?: string;
	receivedCallbackUrl?: string;
	receivedAndDeliveryCallbackUrl?: string;
	disconnectedCallbackUrl?: string;
	connectedCallbackUrl?: string;
	messageStatusCallbackUrl?: string;
	presenceChatCallbackUrl?: string;
	notifySentByMe?: boolean;
}

/**
 * Raw instance data from API (matches Go JSON tags from partners.go)
 *
 * The Go API returns both short names (id, token, created) and full names (instanceId, instanceToken)
 */
export interface InstanceData {
	// Primary fields from API
	readonly id: string;
	readonly token: string;
	readonly instanceId: string;
	readonly instanceToken: string;
	name: string;
	sessionName: string;
	/** Creation date as ISO string (from Go: Created) */
	readonly created?: string;
	/** Subscription due date as Unix timestamp in milliseconds */
	due?: number;
	phoneConnected: boolean;
	whatsappConnected: boolean;
	subscriptionActive: boolean;
	middleware: InstanceMiddleware;
	storeJid?: string;
	isDevice?: boolean;
	businessDevice?: boolean;
	callRejectAuto: boolean;
	callRejectMessage?: string;
	autoReadMessage: boolean;
	canceledAt?: string;
	webhooks?: WebhookSettings;
	connectionStatus?: string;
	lastConnectedAt?: string;
	workerId?: string;
	desiredWorkerId?: string;
	// Webhook URLs flattened at top level
	deliveryCallbackUrl?: string;
	receivedCallbackUrl?: string;
	receivedAndDeliveryCallbackUrl?: string;
	disconnectedCallbackUrl?: string;
	connectedCallbackUrl?: string;
	messageStatusCallbackUrl?: string;
	presenceChatCallbackUrl?: string;
	notifySentByMe?: boolean;
}

/**
 * Instance type - same as InstanceData since API already returns all fields
 */
export type Instance = InstanceData;

/**
 * Normalize instance data from API
 * The API already returns all required fields, so this is a pass-through
 */
export function normalizeInstance(data: InstanceData): Instance {
	return data;
}

/**
 * Normalize an array of instances
 */
export function normalizeInstances(data: InstanceData[]): Instance[] {
	return data ?? [];
}

/**
 * Instance connection status
 * Real-time connection state information
 */
export interface InstanceStatus {
	/** Overall connection status */
	connected: boolean;

	/** Error message if connection failed */
	error: string;

	/** Physical device connection status */
	smartphoneConnected: boolean;
}

/**
 * Device information
 * Detailed information about the connected WhatsApp device
 */
export interface DeviceInfo {
	/** Phone number in international format */
	phone: string;

	/** Profile picture URL */
	imgUrl?: string;

	/** Profile display name */
	name: string;

	/** Device technical details */
	device: DeviceDetails;

	/** Original device identifier */
	originalDevice?: string;

	/** Session identifier */
	sessionId?: number;

	/** WhatsApp Business account indicator */
	isBusiness: boolean;
}

/**
 * Device technical specifications
 */
export interface DeviceDetails {
	/** Session name identifier */
	sessionName: string;

	/** Device model name */
	device_model: string;

	/** WhatsApp version */
	wa_version: string;

	/** Operating system platform */
	platform: string;

	/** Operating system version */
	os_version: string;

	/** Device manufacturer */
	device_manufacturer: string;

	/** Mobile Country Code */
	mcc?: string;

	/** Mobile Network Code */
	mnc?: string;

	/** OS build number */
	osbuildnumber?: string;
}

/**
 * Paginated instance list response
 * Content is normalized with UI-friendly aliases (id, token, created)
 */
export interface InstanceListResponse {
	/** Total number of instances */
	total: number;

	/** Total number of pages */
	totalPage: number;

	/** Items per page */
	pageSize: number;

	/** Current page number (1-indexed) */
	page: number;

	/** Instance array for current page (normalized with aliases) */
	content: Instance[];
}

/**
 * Raw paginated instance list response from API
 */
export interface RawInstanceListResponse {
	total: number;
	totalPage: number;
	pageSize: number;
	page: number;
	content: InstanceData[];
}

/**
 * Request payload for creating a new instance
 */
export interface CreateInstanceRequest {
	/** Instance display name (required) */
	name: string;

	/** WhatsApp session identifier */
	sessionName?: string;

	// Webhook URLs
	deliveryCallbackUrl?: string;
	receivedCallbackUrl?: string;
	receivedAndDeliveryCallbackUrl?: string;
	messageStatusCallbackUrl?: string;
	connectedCallbackUrl?: string;
	disconnectedCallbackUrl?: string;
	presenceChatCallbackUrl?: string;

	// Settings
	notifySentByMe?: boolean;
	callRejectAuto?: boolean;
	callRejectMessage?: string;
	autoReadMessage?: boolean;

	// Device Configuration
	/** Use device-based connection */
	isDevice?: boolean;

	/** Configure as WhatsApp Business */
	businessDevice?: boolean;
}

/**
 * Response after creating a new instance
 */
export interface CreateInstanceResponse {
	/** Unique instance identifier (UUID) */
	readonly instanceId: string;

	/** Authentication token */
	readonly instanceToken: string;

	/** Subscription expiration timestamp */
	due?: string;

	/** Instance name */
	name: string;

	/** Session identifier */
	sessionName: string;

	/** Alias for instanceId (for UI compatibility) */
	readonly id: string;

	/** Alias for instanceToken (for UI compatibility) */
	readonly token: string;
}

/**
 * QR code response for web pairing
 */
export interface QRCodeResponse {
	/** Base64-encoded PNG image with data URI prefix */
	image: string;
}

/**
 * Phone pairing code response
 */
export interface PhonePairingResponse {
	/** 6-digit pairing code in XXX-XXX format */
	code: string;
}

/**
 * Instance configuration settings
 */
export interface InstanceSettings {
	/** Auto-reject incoming calls */
	callRejectAuto: boolean;

	/** Custom rejection message */
	callRejectMessage?: string;

	/** Auto-read incoming messages */
	autoReadMessage: boolean;

	/** Send webhooks for own messages */
	notifySentByMe: boolean;
}

/**
 * Type guard to check if instance is connected
 */
export function isInstanceConnected(instance: InstanceData): boolean {
	return instance.whatsappConnected && instance.phoneConnected;
}

/**
 * Type guard to check if instance has active subscription
 */
export function hasActiveSubscription(instance: InstanceData): boolean {
	if (!instance.subscriptionActive) return false;
	if (!instance.due) return instance.subscriptionActive;
	return new Date(instance.due).getTime() > Date.now();
}

/**
 * Extract instance identifier
 */
export function getInstanceId(instance: InstanceData | Instance): string {
	return instance.instanceId;
}

/**
 * Extract instance token
 */
export function getInstanceToken(instance: InstanceData | Instance): string {
	return instance.instanceToken;
}
