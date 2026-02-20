// =============================================================================
// SICREDI DUAL CLIENT
// =============================================================================
// Duas APIs separadas, dois fluxos de autenticação:
//
// 1. API PIX (mTLS) — api-pix.sicredi.com.br
//    - Auth: OAuth2 client_credentials + Basic Auth
//    - TLS: Certificado digital (mTLS) via Bun.file()
//    - Endpoints BACEN padrão: /api/v2/cob, /api/v2/cobv, /api/v2/webhook
//    - Usado por: pix.ts (cob), webhook.ts
//
// 2. API Parceiro (OpenAPI Gateway) — api-parceiro.sicredi.com.br
//    - Auth: OAuth2 password grant + x-api-key header
//    - TLS: HTTPS padrão (sem mTLS)
//    - Endpoints Sicredi: /cobranca/boleto/v1/boletos (boleto híbrido + consulta)
//    - Usado por: boleto-hibrido.ts (criar boleto, consultar status, PDF)
// =============================================================================

// ---- Shared types ----

interface TokenCache {
	accessToken: string;
	expiresAt: number; // Unix timestamp ms
}

interface ParceiroTokenCache extends TokenCache {
	refreshToken: string;
	refreshExpiresAt: number;
}

// =============================================================================
// 1. PIX CLIENT (mTLS + client_credentials)
// Base URL: https://api-pix.sicredi.com.br
// =============================================================================

const PIX_BASE_URL = "https://api-pix.sicredi.com.br";
const PIX_AUTH_URL = `${PIX_BASE_URL}/oauth/token`;

class SicrediPixClient {
	private tokenCache: TokenCache | null = null;
	private tokenPromise: Promise<string> | null = null;

	private get basicAuth(): string {
		return process.env.SICREDI_PIX_BASIC_AUTH || "";
	}

	private get certPath(): string {
		const path = process.env.SICREDI_CERT_PATH;
		if (!path)
			throw new Error(
				"SICREDI_CERT_PATH env var is required for PIX mTLS",
			);
		return path;
	}

	private get keyPath(): string {
		const path = process.env.SICREDI_KEY_PATH;
		if (!path)
			throw new Error(
				"SICREDI_KEY_PATH env var is required for PIX mTLS",
			);
		return path;
	}

	private get chainPath(): string | undefined {
		return process.env.SICREDI_CHAIN_PATH;
	}

	/**
	 * Make a fetch request with mTLS client certificates.
	 * Uses Bun's native `tls` option on fetch — this is Bun-only and will
	 * not work in Node.js or other runtimes.
	 */
	private mtlsFetch(url: string, init: RequestInit): Promise<Response> {
		const tlsConfig: Record<string, unknown> = {
			key: Bun.file(this.keyPath),
			cert: Bun.file(this.certPath),
		};

		if (this.chainPath) {
			tlsConfig.ca = Bun.file(this.chainPath);
		}

		return fetch(url, {
			...init,
			tls: tlsConfig,
		});
	}

	/**
	 * Get OAuth2 access token (client_credentials via API PIX mTLS).
	 * Caches token in memory. Thread-safe: concurrent callers share the same promise.
	 */
	async getToken(): Promise<string> {
		if (
			this.tokenCache &&
			Date.now() < this.tokenCache.expiresAt - 30_000
		) {
			return this.tokenCache.accessToken;
		}

		if (this.tokenPromise) return this.tokenPromise;

		this.tokenPromise = this._fetchToken();
		try {
			return await this.tokenPromise;
		} finally {
			this.tokenPromise = null;
		}
	}

	private async _fetchToken(): Promise<string> {
		const body = new URLSearchParams({
			grant_type: "client_credentials",
			scope: "cob.read cob.write webhook.read webhook.write",
		});

		const response = await this.mtlsFetch(PIX_AUTH_URL, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
				Accept: "application/json",
				Authorization: `Basic ${this.basicAuth}`,
			},
			body: body.toString(),
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(
				`Sicredi PIX token error (${response.status}): ${errorText}`,
			);
		}

		const data = (await response.json()) as {
			access_token: string;
			expires_in: number;
			token_type: string;
			scope: string;
		};

		this.tokenCache = {
			accessToken: data.access_token,
			expiresAt: Date.now() + data.expires_in * 1000,
		};

		return data.access_token;
	}

	/**
	 * Make an authenticated request to the Sicredi PIX API (mTLS).
	 * Automatically retries once on 401 (expired token).
	 */
	async request<T>(
		method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
		path: string,
		body?: unknown,
	): Promise<T> {
		const attempt = async (retry: boolean): Promise<T> => {
			const token = await this.getToken();

			const response = await this.mtlsFetch(`${PIX_BASE_URL}${path}`, {
				method,
				headers: {
					Authorization: `Bearer ${token}`,
					"Content-Type": "application/json",
					Accept: "application/json",
				},
				body: body ? JSON.stringify(body) : undefined,
			});

			if (response.status === 401 && retry) {
				this.tokenCache = null;
				return attempt(false);
			}

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(
					`Sicredi PIX API error (${response.status} ${method} ${path}): ${errorText}`,
				);
			}

			if (response.status === 204) return {} as T;
			return (await response.json()) as T;
		};

		return attempt(true);
	}
}

// =============================================================================
// 2. PARCEIRO CLIENT (API Key + password grant)
// Base URL: https://api-parceiro.sicredi.com.br
// =============================================================================

const PARCEIRO_AUTH_URL =
	process.env.SICREDI_PARCEIRO_AUTH_URL ||
	"https://api-parceiro.sicredi.com.br/auth/openapi/token";

function getParceiroBaseUrl(): string {
	return (
		process.env.SICREDI_PARCEIRO_BASE_URL ||
		"https://api-parceiro.sicredi.com.br"
	);
}

class SicrediParceiroClient {
	private tokenData: ParceiroTokenCache | null = null;
	private tokenPromise: Promise<string> | null = null;

	private get apiKey(): string {
		return process.env.SICREDI_PARCEIRO_API_KEY || "";
	}

	private get username(): string {
		return process.env.SICREDI_PARCEIRO_USERNAME || "";
	}

	private get password(): string {
		return process.env.SICREDI_PARCEIRO_PASSWORD || "";
	}

	private get cooperativa(): string {
		return process.env.SICREDI_PARCEIRO_COOPERATIVA || "";
	}

	private get posto(): string {
		return process.env.SICREDI_PARCEIRO_POSTO || "";
	}

	private get codigoBeneficiario(): string {
		return process.env.SICREDI_BENEFICIARIO_CNPJ || "";
	}

	/**
	 * Get OAuth2 access token (password grant via API Parceiro).
	 * Caches token in memory. Uses refresh token when possible.
	 * Thread-safe: concurrent callers share the same promise.
	 */
	async getToken(): Promise<string> {
		if (this.tokenData && Date.now() < this.tokenData.expiresAt - 30_000) {
			return this.tokenData.accessToken;
		}

		if (this.tokenPromise) return this.tokenPromise;

		// Try refresh if refresh token is still valid
		if (
			this.tokenData &&
			this.tokenData.refreshToken &&
			Date.now() < this.tokenData.refreshExpiresAt - 30_000
		) {
			this.tokenPromise = this._refreshToken(this.tokenData.refreshToken);
		} else {
			this.tokenPromise = this._fetchToken();
		}

		try {
			return await this.tokenPromise;
		} finally {
			this.tokenPromise = null;
		}
	}

	private async _fetchToken(): Promise<string> {
		const body = new URLSearchParams({
			grant_type: "password",
			username: this.username,
			password: this.password,
			scope: "cobranca",
		});

		const response = await fetch(PARCEIRO_AUTH_URL, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
				Accept: "application/json",
				"x-api-key": this.apiKey,
				context: "COBRANCA",
			},
			body: body.toString(),
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(
				`Sicredi Parceiro token error (${response.status}): ${errorText}`,
			);
		}

		const data = (await response.json()) as {
			access_token: string;
			expires_in: number;
			refresh_token: string;
			refresh_expires_in: number;
			token_type: string;
		};

		this.tokenData = {
			accessToken: data.access_token,
			refreshToken: data.refresh_token,
			expiresAt: Date.now() + data.expires_in * 1000,
			refreshExpiresAt: Date.now() + data.refresh_expires_in * 1000,
		};

		return this.tokenData.accessToken;
	}

	private async _refreshToken(refreshToken: string): Promise<string> {
		const body = new URLSearchParams({
			grant_type: "refresh_token",
			refresh_token: refreshToken,
		});

		const response = await fetch(PARCEIRO_AUTH_URL, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
				Accept: "application/json",
				"x-api-key": this.apiKey,
				context: "COBRANCA",
			},
			body: body.toString(),
		});

		if (!response.ok) {
			// Refresh failed — do full re-auth
			this.tokenData = null;
			return this._fetchToken();
		}

		const data = (await response.json()) as {
			access_token: string;
			expires_in: number;
			refresh_token: string;
			refresh_expires_in: number;
			token_type: string;
		};

		this.tokenData = {
			accessToken: data.access_token,
			refreshToken: data.refresh_token,
			expiresAt: Date.now() + data.expires_in * 1000,
			refreshExpiresAt: Date.now() + data.refresh_expires_in * 1000,
		};

		return this.tokenData.accessToken;
	}

	/** Expose codigoBeneficiario for boleto creation */
	getCodigoBeneficiario(): string {
		return this.codigoBeneficiario;
	}

	/**
	 * Make an authenticated request to the Sicredi API Parceiro.
	 * Automatically retries once on 401 (expired token).
	 * Includes cooperativa and posto headers for cobrança endpoints.
	 */
	async request<T>(
		method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
		path: string,
		body?: unknown,
	): Promise<T> {
		const attempt = async (retry: boolean): Promise<T> => {
			const token = await this.getToken();
			const baseUrl = getParceiroBaseUrl();

			const headers: Record<string, string> = {
				Authorization: `Bearer ${token}`,
				"Content-Type": "application/json",
				Accept: "application/json",
				"x-api-key": this.apiKey,
				context: "COBRANCA",
				cooperativa: this.cooperativa,
				posto: this.posto,
			};

			const response = await fetch(`${baseUrl}${path}`, {
				method,
				headers,
				body: body ? JSON.stringify(body) : undefined,
			});

			if (response.status === 401 && retry) {
				this.tokenData = null;
				return attempt(false);
			}

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(
					`Sicredi Parceiro API error (${response.status} ${method} ${path}): ${errorText}`,
				);
			}

			if (response.status === 204) return {} as T;
			return (await response.json()) as T;
		};

		return attempt(true);
	}
}

// =============================================================================
// EXPORTS
// =============================================================================

/** PIX operations (cob, webhook) — mTLS via api-pix.sicredi.com.br */
export const sicrediPix = new SicrediPixClient();

/** Boleto Híbrido operations (cobv) — API Parceiro via api-parceiro.sicredi.com.br */
export const sicrediParceiro = new SicrediParceiroClient();
