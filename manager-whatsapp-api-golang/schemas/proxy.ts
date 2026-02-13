import { z } from "zod";

/**
 * Proxy URL validation pattern.
 * Accepts http://, https://, or socks5:// URLs.
 */
const proxyUrlPattern = /^(https?|socks5):\/\/.+/;

/**
 * Schema for proxy URL validation.
 * Must be a valid URL with http, https, or socks5 scheme.
 */
export const ProxyUrlSchema = z
	.string()
	.min(1, "Proxy URL is required")
	.regex(proxyUrlPattern, "Must be http://, https://, or socks5:// URL")
	.refine(
		(url) => {
			try {
				new URL(url);
				return true;
			} catch {
				return false;
			}
		},
		{ message: "Invalid URL format" },
	);

/**
 * Schema for proxy configuration form.
 */
export const ProxyConfigSchema = z.object({
	proxyUrl: ProxyUrlSchema,
	noWebsocket: z.boolean().default(false),
	onlyLogin: z.boolean().default(false),
	noMedia: z.boolean().default(false),
});

/**
 * Schema for proxy test form.
 */
export const ProxyTestSchema = z.object({
	proxyUrl: ProxyUrlSchema,
});

/**
 * Schema for proxy swap form.
 */
export const ProxySwapSchema = z.object({
	proxyUrl: ProxyUrlSchema,
});

// Inferred TypeScript types
export type ProxyConfigInput = z.input<typeof ProxyConfigSchema>;
export type ProxyConfigFormValues = z.infer<typeof ProxyConfigSchema>;
export type ProxyTestInput = z.input<typeof ProxyTestSchema>;
export type ProxySwapInput = z.input<typeof ProxySwapSchema>;
