import { z } from "zod";
import { createLogger } from "@/lib/logger";

const log = createLogger("env");

const envSchema = z.object({
	// Database
	DATABASE_URL: z.string().url(),

	// Better Auth
	BETTER_AUTH_SECRET: z.string().min(32),
	BETTER_AUTH_URL: z.string().url(),
	BETTER_AUTH_TRUSTED_ORIGINS: z.string().default("http://localhost:3000"),

	// ZedaAPI Integration
	ZEDAAPI_BASE_URL: z.string().url(),
	ZEDAAPI_PARTNER_TOKEN: z.string().min(16),
	ZEDAAPI_CLIENT_TOKEN: z.string().min(16),
	ZEDAAPI_WEBHOOK_SECRET: z.string().min(1),

	// Stripe
	STRIPE_SECRET_KEY: z.string().startsWith("sk_"),
	STRIPE_WEBHOOK_SECRET: z.string().startsWith("whsec_"),
	NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY: z.string().startsWith("pk_"),

	// Sicredi PIX
	SICREDI_PIX_BASIC_AUTH: z.string().optional(),
	SICREDI_PIX_KEY: z.string().optional(),
	SICREDI_CERT_PATH: z.string().optional(),
	SICREDI_KEY_PATH: z.string().optional(),
	SICREDI_CHAIN_PATH: z.string().optional(),
	SICREDI_WEBHOOK_SECRET: z.string().optional(),

	// Sicredi Parceiro (Boleto)
	SICREDI_PARCEIRO_AUTH_URL: z.string().url().optional(),
	SICREDI_PARCEIRO_BASE_URL: z.string().url().optional(),
	SICREDI_PARCEIRO_API_KEY: z.string().optional(),
	SICREDI_PARCEIRO_USERNAME: z.string().optional(),
	SICREDI_PARCEIRO_PASSWORD: z.string().optional(),
	SICREDI_PARCEIRO_COOPERATIVA: z.string().optional(),
	SICREDI_PARCEIRO_POSTO: z.string().optional(),
	SICREDI_BENEFICIARIO_CNPJ: z.string().optional(),
	SICREDI_BENEFICIARIO_NOME: z.string().optional(),

	// Redis
	REDIS_URL: z.string().default("redis://localhost:6379"),

	// S3/MinIO
	S3_ENDPOINT: z.string().optional(),
	S3_REGION: z.string().default("us-east-1"),
	S3_ACCESS_KEY_ID: z.string().optional(),
	S3_SECRET_ACCESS_KEY: z.string().optional(),
	S3_BUCKET: z.string().optional(),
	S3_PUBLIC_URL: z.string().optional(),
	S3_PATH_STYLE: z.string().default("true"),

	// Encryption
	ENCRYPTION_KEY: z.string().length(64),

	// Email
	SMTP_HOST: z.string().optional(),
	SMTP_PORT: z.coerce.number().default(587),
	SMTP_USER: z.string().optional(),
	SMTP_PASSWORD: z.string().optional(),
	SMTP_FROM: z.string().email().optional(),

	// Worker
	WORKER_LOG_LEVEL: z
		.enum(["DEBUG", "INFO", "WARN", "ERROR"])
		.default("INFO"),

	// App
	NODE_ENV: z
		.enum(["development", "production", "test"])
		.default("development"),
	NEXT_PUBLIC_APP_URL: z.string().url().default("http://localhost:3000"),
	ADMIN_EMAIL: z.string().email().optional(),
	ADMIN_PASSWORD: z.string().optional(),

	// Contact / Landing Page
	CONTACT_WEBHOOK_URL: z.string().url().optional(),
	NEXT_PUBLIC_CONTACT_WHATSAPP: z.string().optional(),

	// Content API
	CONTENT_API_KEY: z.string().min(32).optional(),

	// Waitlist
	WAITLIST_ENABLED: z.string().default("true"),
	WAITLIST_REQUIRE_APPROVAL: z.string().default("true"),
});

export function getEnv() {
	const parsed = envSchema.safeParse(process.env);
	if (!parsed.success) {
		log.error("Invalid environment variables", {
			fieldErrors: parsed.error.flatten().fieldErrors,
		});
		throw new Error("Invalid environment variables");
	}
	return parsed.data;
}

export const clientEnv = {
	NEXT_PUBLIC_APP_URL:
		process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000",
	NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY:
		process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || "",
	NEXT_PUBLIC_CONTACT_WHATSAPP:
		process.env.NEXT_PUBLIC_CONTACT_WHATSAPP || "",
};

export type Env = z.infer<typeof envSchema>;
