export class AppError extends Error {
	public readonly statusCode: number;
	public readonly code: string;

	constructor(message: string, statusCode = 500, code = "INTERNAL_ERROR") {
		super(message);
		this.name = "AppError";
		this.statusCode = statusCode;
		this.code = code;
	}
}

export class AuthError extends AppError {
	constructor(message = "Authentication required") {
		super(message, 401, "AUTH_ERROR");
		this.name = "AuthError";
	}
}

export class NotFoundError extends AppError {
	constructor(resource = "Resource") {
		super(`${resource} not found`, 404, "NOT_FOUND");
		this.name = "NotFoundError";
	}
}

export class ForbiddenError extends AppError {
	constructor(message = "Access denied") {
		super(message, 403, "FORBIDDEN");
		this.name = "ForbiddenError";
	}
}

export class ValidationError extends AppError {
	public readonly errors: Record<string, string[]>;

	constructor(errors: Record<string, string[]>) {
		super("Validation failed", 422, "VALIDATION_ERROR");
		this.name = "ValidationError";
		this.errors = errors;
	}
}

export class PaymentError extends AppError {
	public readonly provider: string;

	constructor(message: string, provider: string) {
		super(message, 402, "PAYMENT_ERROR");
		this.name = "PaymentError";
		this.provider = provider;
	}
}

export class ZedaAPIError extends AppError {
	public readonly upstream: string | undefined;

	constructor(message: string, statusCode = 502, upstream?: string) {
		super(message, statusCode, "ZEDAAPI_ERROR");
		this.name = "ZedaAPIError";
		this.upstream = upstream;
	}
}

export class RateLimitError extends AppError {
	public readonly retryAfterMs: number;

	constructor(retryAfterMs: number) {
		super("Too many requests", 429, "RATE_LIMIT");
		this.name = "RateLimitError";
		this.retryAfterMs = retryAfterMs;
	}
}
