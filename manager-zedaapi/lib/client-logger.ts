/**
 * Client-side Logger — for "use client" components (browser).
 *
 * Lightweight wrapper around console with scoped prefixes.
 * Mirrors the server Logger interface so usage is consistent.
 *
 * Usage:
 *   import { createClientLogger } from "@/lib/client-logger";
 *   const log = createClientLogger("billing");
 *   log.error("Payment failed", { invoiceId });
 */

export interface ClientLogger {
	debug(message: string, ctx?: Record<string, unknown>): void;
	info(message: string, ctx?: Record<string, unknown>): void;
	warn(message: string, ctx?: Record<string, unknown>): void;
	error(message: string, ctx?: Record<string, unknown>): void;
}

export function createClientLogger(scope: string): ClientLogger {
	const fmt = (msg: string, ctx?: Record<string, unknown>) =>
		ctx && Object.keys(ctx).length > 0
			? `[${scope}] ${msg} ${JSON.stringify(ctx)}`
			: `[${scope}] ${msg}`;

	return {
		debug: (msg, ctx) => console.debug(fmt(msg, ctx)),
		info: (msg, ctx) => console.log(fmt(msg, ctx)),
		warn: (msg, ctx) => console.warn(fmt(msg, ctx)),
		error: (msg, ctx) => console.error(fmt(msg, ctx)),
	};
}
