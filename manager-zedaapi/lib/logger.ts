/**
 * Structured Logger — shared across Manager (Next.js) and Workers (BullMQ).
 *
 * Features:
 * - ISO timestamps on every line
 * - Log levels: DEBUG, INFO, WARN, ERROR
 * - Duration tracking via timer()
 * - Structured context (key=value pairs)
 * - Scoped loggers via child()
 * - JSON output in production (for log aggregators)
 * - Human-readable output in development
 *
 * Usage:
 *   import { createLogger } from "@/lib/logger";
 *   const log = createLogger("service:billing");
 *   log.info("Payment processed", { userId, amount: 29 });
 *   log.error("Payment failed", { userId, error: err.message });
 *
 *   const done = log.timer("stripe-checkout");
 *   await createCheckoutSession();
 *   done(); // logs: "stripe-checkout completed | duration=1.2s"
 */

export type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR";

const LOG_LEVEL_PRIORITY: Record<LogLevel, number> = {
	DEBUG: 0,
	INFO: 1,
	WARN: 2,
	ERROR: 3,
};

function getMinLevel(): LogLevel {
	// Workers use WORKER_LOG_LEVEL, server uses LOG_LEVEL
	const env =
		(
			process.env.WORKER_LOG_LEVEL ?? process.env.LOG_LEVEL
		)?.toUpperCase() ?? "";
	if (env in LOG_LEVEL_PRIORITY) return env as LogLevel;
	return process.env.NODE_ENV === "production" ? "INFO" : "DEBUG";
}

function shouldLog(level: LogLevel): boolean {
	return LOG_LEVEL_PRIORITY[level] >= LOG_LEVEL_PRIORITY[getMinLevel()];
}

function isProduction(): boolean {
	return process.env.NODE_ENV === "production";
}

// ── Formatters ──────────────────────────

function formatDuration(ms: number): string {
	if (ms < 1000) return `${ms}ms`;
	if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`;
	return `${(ms / 60_000).toFixed(1)}min`;
}

function formatContextText(ctx?: Record<string, unknown>): string {
	if (!ctx || Object.keys(ctx).length === 0) return "";
	const parts = Object.entries(ctx)
		.filter(([, v]) => v !== undefined && v !== null)
		.map(
			([k, v]) => `${k}=${typeof v === "object" ? JSON.stringify(v) : v}`,
		);
	return parts.length > 0 ? ` | ${parts.join(", ")}` : "";
}

// ── Core ──────────────────────────

function emit(
	level: LogLevel,
	scope: string,
	message: string,
	ctx?: Record<string, unknown>,
): void {
	if (!shouldLog(level)) return;

	const timestamp = new Date().toISOString();

	let line: string;
	if (isProduction()) {
		// JSON lines — parseable by Datadog, ELK, CloudWatch, Loki
		line = JSON.stringify({
			ts: timestamp,
			level,
			scope,
			msg: message,
			...ctx,
		});
	} else {
		// Human-readable — for local development
		const contextStr = formatContextText(ctx);
		line = `${timestamp} [${level}] [${scope}] ${message}${contextStr}`;
	}

	switch (level) {
		case "ERROR":
			console.error(line);
			break;
		case "WARN":
			console.warn(line);
			break;
		default:
			console.log(line);
	}
}

// ── Public Interface ──────────────────────────

export interface Logger {
	debug(message: string, ctx?: Record<string, unknown>): void;
	info(message: string, ctx?: Record<string, unknown>): void;
	warn(message: string, ctx?: Record<string, unknown>): void;
	error(message: string, ctx?: Record<string, unknown>): void;
	/** Start a timer. Call the returned function to log completion + duration. */
	timer(label: string, ctx?: Record<string, unknown>): () => void;
	/** Create a child logger with a sub-scope (e.g. "billing:checkout"). */
	child(subscope: string): Logger;
}

export function createLogger(scope: string): Logger {
	return {
		debug: (msg, ctx) => emit("DEBUG", scope, msg, ctx),
		info: (msg, ctx) => emit("INFO", scope, msg, ctx),
		warn: (msg, ctx) => emit("WARN", scope, msg, ctx),
		error: (msg, ctx) => emit("ERROR", scope, msg, ctx),

		timer(label: string, ctx?: Record<string, unknown>) {
			const start = Date.now();
			emit("DEBUG", scope, `${label} started`, ctx);
			return () => {
				const elapsed = Date.now() - start;
				emit("INFO", scope, `${label} completed`, {
					...ctx,
					duration: formatDuration(elapsed),
				});
			};
		},

		child(subscope: string): Logger {
			return createLogger(`${scope}:${subscope}`);
		},
	};
}
