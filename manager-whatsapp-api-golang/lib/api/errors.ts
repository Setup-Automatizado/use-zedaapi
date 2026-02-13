/**
 * API error classes for WhatsApp API client
 *
 * @module lib/api/errors
 */

/**
 * Custom error class for API failures
 * Extends native Error with HTTP status information and structured error body
 */
export class ApiError extends Error {
	/**
	 * Creates an API error instance
	 *
	 * @param statusCode - HTTP status code
	 * @param statusText - HTTP status text (e.g., "Not Found")
	 * @param body - Response body (may contain error details)
	 */
	constructor(
		public readonly statusCode: number,
		public readonly statusText: string,
		public readonly body: unknown,
	) {
		super(`API Error: ${statusCode} ${statusText}`);
		this.name = "ApiError";

		// Maintains proper stack trace for where error was thrown (V8 only)
		if (Error.captureStackTrace) {
			Error.captureStackTrace(this, ApiError);
		}
	}

	/**
	 * Checks if error is a 404 Not Found
	 */
	get isNotFound(): boolean {
		return this.statusCode === 404;
	}

	/**
	 * Checks if error is a 401 Unauthorized
	 */
	get isUnauthorized(): boolean {
		return this.statusCode === 401;
	}

	/**
	 * Checks if error is a 403 Forbidden
	 */
	get isForbidden(): boolean {
		return this.statusCode === 403;
	}

	/**
	 * Checks if error is a 5xx Server Error
	 */
	get isServerError(): boolean {
		return this.statusCode >= 500;
	}

	/**
	 * Checks if error is a 4xx Client Error
	 */
	get isClientError(): boolean {
		return this.statusCode >= 400 && this.statusCode < 500;
	}

	/**
	 * Extracts human-readable error message from response body
	 * Falls back to status code and text if no structured error is available
	 */
	override get message(): string {
		// Try to extract error from structured response
		if (this.body && typeof this.body === "object" && "error" in this.body) {
			const errorField = (this.body as Record<string, unknown>).error;
			if (typeof errorField === "string") {
				return errorField;
			}
		}

		// Try to extract message field
		if (this.body && typeof this.body === "object" && "message" in this.body) {
			const messageField = (this.body as Record<string, unknown>).message;
			if (typeof messageField === "string") {
				return messageField;
			}
		}

		// Fallback to status code and text
		return `${this.statusCode} ${this.statusText}`;
	}

	/**
	 * Serializes error to plain object for logging/debugging
	 */
	toJSON(): Record<string, unknown> {
		return {
			name: this.name,
			message: this.message,
			statusCode: this.statusCode,
			statusText: this.statusText,
			body: this.body,
			stack: this.stack,
		};
	}
}

/**
 * Type guard to check if error is an ApiError instance
 */
export function isApiError(error: unknown): error is ApiError {
	return error instanceof ApiError;
}
