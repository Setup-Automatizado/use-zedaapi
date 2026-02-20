import { ZedaAPIError } from "@/lib/errors";
import type {
	CreateInstanceRequest,
	CreateInstanceResponse,
	DeviceInfo,
	InstanceStatusResponse,
	ListInstancesParams,
	PaginatedResponse,
	PhoneCodeResponse,
	QRCodeResponse,
	WebhookAllUpdateRequest,
	ZedaAPIErrorResponse,
	ZedaAPIInstance,
} from "@/types/zedaapi";

// =============================================================================
// Configuration
// =============================================================================

const DEFAULT_TIMEOUT_MS = 30_000;
const MAX_RETRIES = 3;
const RETRY_BASE_MS = 1_000;
const CIRCUIT_BREAKER_THRESHOLD = 5;
const CIRCUIT_BREAKER_RESET_MS = 30_000;

// =============================================================================
// ZedaAPI HTTP Client
// =============================================================================

class ZedaAPIClient {
	private readonly baseUrl: string;
	private readonly partnerToken: string;
	private readonly clientToken: string;

	// Circuit breaker state
	private failures = 0;
	private circuitOpenUntil = 0;

	constructor(config: {
		baseUrl: string;
		partnerToken: string;
		clientToken: string;
	}) {
		this.baseUrl = config.baseUrl.replace(/\/+$/, "");
		this.partnerToken = config.partnerToken;
		this.clientToken = config.clientToken;
	}

	// -------------------------------------------------------------------------
	// Partner endpoints (Bearer auth)
	// -------------------------------------------------------------------------

	/** POST /instances/integrator/on-demand */
	async createInstance(
		data: CreateInstanceRequest,
	): Promise<CreateInstanceResponse> {
		return this.partnerRequest<CreateInstanceResponse>(
			"/instances/integrator/on-demand",
			{ method: "POST", body: JSON.stringify(data) },
		);
	}

	/** POST /instances/{id}/token/{token}/integrator/on-demand/subscription */
	async activateSubscription(
		instanceId: string,
		token: string,
	): Promise<void> {
		await this.partnerRequest<void>(
			`/instances/${instanceId}/token/${token}/integrator/on-demand/subscription`,
			{ method: "POST" },
		);
	}

	/** POST /instances/{id}/token/{token}/integrator/on-demand/cancel */
	async cancelSubscription(instanceId: string, token: string): Promise<void> {
		await this.partnerRequest<void>(
			`/instances/${instanceId}/token/${token}/integrator/on-demand/cancel`,
			{ method: "POST" },
		);
	}

	/** DELETE /instances/{id} */
	async deleteInstance(instanceId: string): Promise<void> {
		await this.partnerRequest<void>(`/instances/${instanceId}`, {
			method: "DELETE",
		});
	}

	/** GET /instances */
	async listInstances(
		params?: ListInstancesParams,
	): Promise<PaginatedResponse<ZedaAPIInstance>> {
		const searchParams = new URLSearchParams();
		if (params?.page) searchParams.set("page", String(params.page));
		if (params?.pageSize)
			searchParams.set("pageSize", String(params.pageSize));
		const qs = searchParams.toString();
		const path = qs ? `/instances?${qs}` : "/instances";
		return this.partnerRequest<PaginatedResponse<ZedaAPIInstance>>(path, {
			method: "GET",
		});
	}

	// -------------------------------------------------------------------------
	// Instance endpoints (Client-Token auth)
	// Route pattern: /instances/{id}/token/{token}/...
	// -------------------------------------------------------------------------

	/** GET /instances/{id}/token/{token}/status */
	async getStatus(
		instanceId: string,
		token: string,
	): Promise<InstanceStatusResponse> {
		return this.instanceRequest<InstanceStatusResponse>(
			instanceId,
			token,
			"/status",
			{ method: "GET" },
		);
	}

	/** GET /instances/{id}/token/{token}/qr-code */
	async getQRCode(
		instanceId: string,
		token: string,
	): Promise<QRCodeResponse> {
		return this.instanceRequest<QRCodeResponse>(
			instanceId,
			token,
			"/qr-code",
			{ method: "GET" },
		);
	}

	/** GET /instances/{id}/token/{token}/device */
	async getDevice(instanceId: string, token: string): Promise<DeviceInfo> {
		return this.instanceRequest<DeviceInfo>(instanceId, token, "/device", {
			method: "GET",
		});
	}

	/** POST /instances/{id}/token/{token}/restart */
	async restart(instanceId: string, token: string): Promise<void> {
		await this.instanceRequest<void>(instanceId, token, "/restart", {
			method: "POST",
		});
	}

	/** POST /instances/{id}/token/{token}/disconnect */
	async disconnect(instanceId: string, token: string): Promise<void> {
		await this.instanceRequest<void>(instanceId, token, "/disconnect", {
			method: "POST",
		});
	}

	/** PUT /instances/{id}/token/{token}/update-every-webhooks */
	async updateAllWebhooks(
		instanceId: string,
		token: string,
		data: WebhookAllUpdateRequest,
	): Promise<void> {
		await this.instanceRequest<void>(
			instanceId,
			token,
			"/update-every-webhooks",
			{ method: "PUT", body: JSON.stringify(data) },
		);
	}

	/** GET /instances/{id}/token/{token}/phone-code/{phone} */
	async getPhoneCode(
		instanceId: string,
		token: string,
		phone: string,
	): Promise<PhoneCodeResponse> {
		return this.instanceRequest<PhoneCodeResponse>(
			instanceId,
			token,
			`/phone-code/${encodeURIComponent(phone)}`,
			{ method: "GET" },
		);
	}

	// -------------------------------------------------------------------------
	// Internal: Partner request (Bearer auth)
	// -------------------------------------------------------------------------

	private async partnerRequest<T>(
		path: string,
		init: RequestInit,
	): Promise<T> {
		return this.request<T>(path, {
			...init,
			headers: {
				...(init.headers as Record<string, string>),
				Authorization: `Bearer ${this.partnerToken}`,
				"Content-Type": "application/json",
			},
		});
	}

	// -------------------------------------------------------------------------
	// Internal: Instance request (Client-Token auth)
	// -------------------------------------------------------------------------

	private async instanceRequest<T>(
		instanceId: string,
		token: string,
		subPath: string,
		init: RequestInit,
	): Promise<T> {
		const path = `/instances/${instanceId}/token/${token}${subPath}`;
		return this.request<T>(path, {
			...init,
			headers: {
				...(init.headers as Record<string, string>),
				"Client-Token": this.clientToken,
				"Content-Type": "application/json",
			},
		});
	}

	// -------------------------------------------------------------------------
	// Internal: Core request with retry + circuit breaker
	// -------------------------------------------------------------------------

	private async request<T>(path: string, init: RequestInit): Promise<T> {
		this.checkCircuitBreaker();

		let lastError: Error | null = null;

		for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
			if (attempt > 0) {
				const delay = RETRY_BASE_MS * 2 ** (attempt - 1);
				await sleep(delay);
			}

			try {
				const url = `${this.baseUrl}${path}`;
				const controller = new AbortController();
				const timeout = setTimeout(
					() => controller.abort(),
					DEFAULT_TIMEOUT_MS,
				);

				const response = await fetch(url, {
					...init,
					signal: controller.signal,
				});

				clearTimeout(timeout);

				this.onSuccess();

				if (!response.ok) {
					const errorBody = (await response
						.json()
						.catch(() => null)) as ZedaAPIErrorResponse | null;

					const message =
						errorBody?.error ||
						errorBody?.message ||
						`ZedaAPI returned ${response.status}`;

					// Do not retry 4xx (except 429)
					if (
						response.status >= 400 &&
						response.status < 500 &&
						response.status !== 429
					) {
						throw new ZedaAPIError(message, response.status, path);
					}

					// 429 or 5xx: retry
					lastError = new ZedaAPIError(
						message,
						response.status,
						path,
					);
					this.onFailure();
					continue;
				}

				// 204 No Content or empty body
				const text = await response.text();
				if (!text) return undefined as T;

				return JSON.parse(text) as T;
			} catch (error) {
				if (error instanceof ZedaAPIError) {
					// Already classified as non-retryable 4xx
					if (
						error.statusCode >= 400 &&
						error.statusCode < 500 &&
						error.statusCode !== 429
					) {
						throw error;
					}
					lastError = error;
				} else if (
					error instanceof DOMException &&
					error.name === "AbortError"
				) {
					lastError = new ZedaAPIError("Request timeout", 504, path);
					this.onFailure();
				} else if (error instanceof TypeError) {
					// Network error (DNS, connection refused, etc.)
					lastError = new ZedaAPIError(
						`Network error: ${error.message}`,
						503,
						path,
					);
					this.onFailure();
				} else {
					lastError =
						error instanceof Error
							? error
							: new Error(String(error));
					this.onFailure();
				}
			}
		}

		throw (
			lastError ??
			new ZedaAPIError("Request failed after retries", 502, path)
		);
	}

	// -------------------------------------------------------------------------
	// Circuit breaker
	// -------------------------------------------------------------------------

	private checkCircuitBreaker(): void {
		if (this.failures >= CIRCUIT_BREAKER_THRESHOLD) {
			if (Date.now() < this.circuitOpenUntil) {
				throw new ZedaAPIError(
					"Circuit breaker open: ZedaAPI unavailable",
					503,
				);
			}
			// Half-open: allow one request through
			this.failures = CIRCUIT_BREAKER_THRESHOLD - 1;
		}
	}

	private onSuccess(): void {
		this.failures = 0;
		this.circuitOpenUntil = 0;
	}

	private onFailure(): void {
		this.failures++;
		if (this.failures >= CIRCUIT_BREAKER_THRESHOLD) {
			this.circuitOpenUntil = Date.now() + CIRCUIT_BREAKER_RESET_MS;
		}
	}
}

// =============================================================================
// Helpers
// =============================================================================

function sleep(ms: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

// =============================================================================
// Singleton export
// =============================================================================

function createClient(): ZedaAPIClient {
	const baseUrl = process.env.ZEDAAPI_BASE_URL;
	const partnerToken = process.env.ZEDAAPI_PARTNER_TOKEN;
	const clientToken = process.env.ZEDAAPI_CLIENT_TOKEN;

	if (!baseUrl || !partnerToken || !clientToken) {
		throw new Error(
			"Missing ZedaAPI configuration: ZEDAAPI_BASE_URL, ZEDAAPI_PARTNER_TOKEN, ZEDAAPI_CLIENT_TOKEN",
		);
	}

	return new ZedaAPIClient({ baseUrl, partnerToken, clientToken });
}

const globalForZedaAPI = globalThis as unknown as {
	zedaapiClient: ZedaAPIClient | undefined;
};

export const zedaapi = globalForZedaAPI.zedaapiClient ?? createClient();

if (process.env.NODE_ENV !== "production") {
	globalForZedaAPI.zedaapiClient = zedaapi;
}

export { ZedaAPIClient };
