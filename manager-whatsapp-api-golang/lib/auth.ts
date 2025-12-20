/**
 * Better Auth Configuration
 *
 * Central authentication configuration with email integration,
 * two-factor authentication, social providers, and admin plugin.
 * Implements invitation-only access control.
 *
 * @module lib/auth
 */

import { prismaAdapter } from "better-auth/adapters/prisma";
import { betterAuth } from "better-auth";
import { createAuthMiddleware } from "better-auth/api";
import { twoFactor, admin } from "better-auth/plugins";
import prisma from "@/lib/prisma";
import { ac, adminRole, userRole } from "@/lib/auth/permissions";
import {
	sendPasswordReset,
	sendTwoFactorCode,
	sendTwoFactorEnabled,
	sendTwoFactorDisabled,
	sendLoginAlert,
} from "@/lib/email";

/**
 * Get user agent and IP from request context
 */
function getRequestContextFromHeaders(headers?: Headers) {
	if (!headers) {
		return { device: "Unknown", ipAddress: "Unknown" };
	}

	const userAgent = headers.get("user-agent") || "Unknown";
	const forwardedFor = headers.get("x-forwarded-for");
	const realIp = headers.get("x-real-ip");
	const ipAddress = forwardedFor?.split(",")[0] || realIp || "Unknown";

	// Parse user agent for device info
	let device = userAgent;
	if (userAgent.includes("Chrome")) {
		device = "Chrome Browser";
	} else if (userAgent.includes("Firefox")) {
		device = "Firefox Browser";
	} else if (userAgent.includes("Safari")) {
		device = "Safari Browser";
	} else if (userAgent.includes("Edge")) {
		device = "Edge Browser";
	}

	if (userAgent.includes("Windows")) {
		device += " on Windows";
	} else if (userAgent.includes("Mac")) {
		device += " on macOS";
	} else if (userAgent.includes("Linux")) {
		device += " on Linux";
	} else if (userAgent.includes("iPhone") || userAgent.includes("iPad")) {
		device += " on iOS";
	} else if (userAgent.includes("Android")) {
		device += " on Android";
	}

	return { device, ipAddress };
}

/**
 * Verify if an email is authorized to access the platform.
 * Only emails in the AllowedUser table can login.
 */
async function verifyEmailAuthorization(email: string): Promise<{
	authorized: boolean;
	role: string;
	allowedUserId?: string;
}> {
	const allowedUser = await prisma.allowedUser.findUnique({
		where: { email: email.toLowerCase() },
	});

	if (!allowedUser) {
		return { authorized: false, role: "USER" };
	}

	return {
		authorized: true,
		role: allowedUser.role,
		allowedUserId: allowedUser.id,
	};
}

/**
 * Update user role and mark invitation as accepted
 */
async function markInvitationAccepted(
	userId: string,
	email: string,
	role: string,
): Promise<void> {
	// Update user role
	await prisma.user.update({
		where: { id: userId },
		data: { role },
	});

	// Mark invitation as accepted
	await prisma.allowedUser.update({
		where: { email: email.toLowerCase() },
		data: {
			acceptedAt: new Date(),
			userId: userId,
		},
	});
}

export const auth = betterAuth({
	database: prismaAdapter(prisma, {
		provider: "postgresql",
	}),
	trustedOrigins: [process.env.BETTER_AUTH_URL || "http://localhost:3000"],

	// Email and password authentication
	emailAndPassword: {
		enabled: true,
		requireEmailVerification: false,
		sendResetPassword: async ({ user, url }) => {
			await sendPasswordReset(user.email, {
				userName: user.name || "",
				userEmail: user.email,
				resetUrl: url,
				expiresIn: 60,
			});
		},
	},

	// Social providers
	socialProviders: {
		github: {
			clientId: process.env.GITHUB_CLIENT_ID as string,
			clientSecret: process.env.GITHUB_CLIENT_SECRET as string,
		},
		google: {
			clientId: process.env.GOOGLE_CLIENT_ID as string,
			clientSecret: process.env.GOOGLE_CLIENT_SECRET as string,
			scope: ["openid", "email", "profile"],
		},
	},

	// Plugins
	plugins: [
		twoFactor({
			issuer: "WhatsApp Manager",
			otpOptions: {
				async sendOTP({ user, otp }) {
					await sendTwoFactorCode(user.email, {
						userName: user.name || "",
						userEmail: user.email,
						verificationCode: otp,
						expiresIn: 10,
					});
				},
			},
		}),
		admin({
			ac,
			roles: {
				ADMIN: adminRole,
				USER: userRole,
			},
			defaultRole: "USER",
			adminRoles: ["ADMIN"],
		}),
	],

	// Session configuration
	session: {
		expiresIn: 60 * 60 * 24 * 7, // 7 days
		updateAge: 60 * 60 * 24,
		cookieCache: {
			enabled: true,
			maxAge: 60 * 5,
		},
	},

	// Rate limiting
	rateLimit: {
		window: 60,
		max: 10,
	},

	// Advanced configuration
	advanced: {
		cookiePrefix: "whatsapp-manager",
		// Usar secure cookies apenas se SECURE_COOKIES=true ou se estiver em HTTPS
		// Em homolog com HTTP, precisa ser false para cookies funcionarem
		useSecureCookies: process.env.SECURE_COOKIES === "true",
		defaultCookieAttributes: {
			sameSite: "lax",
			path: "/",
		},
	},

	// Callback URL after OAuth
	callbacks: {
		redirect: {
			afterSignIn: "/dashboard",
			afterSignUp: "/dashboard",
			afterSignOut: "/login",
		},
	},

	// User model customization
	user: {
		additionalFields: {
			role: {
				type: "string",
				defaultValue: "USER",
				input: false,
			},
			banned: {
				type: "boolean",
				defaultValue: false,
				input: false,
			},
			banReason: {
				type: "string",
				required: false,
				input: false,
			},
			banExpiresAt: {
				type: "date",
				required: false,
				input: false,
			},
		},
	},

	// Hooks for intercepting requests (used for notifications)
	hooks: {
		after: createAuthMiddleware(async (ctx) => {
			// Send login alert on successful sign-in
			if (
				ctx.path === "/sign-in/email" ||
				ctx.path === "/sign-in/social"
			) {
				const returned = ctx.context.returned as
					| Record<string, unknown>
					| undefined;
				// Check if sign-in was successful (returns user/session data)
				if (returned && ("user" in returned || "session" in returned)) {
					// Pegar usuario do returned (session ainda nao existe no contexto)
					const user =
						(returned.user as Record<string, unknown>) ||
						((returned.session as Record<string, unknown>)
							?.user as Record<string, unknown>);

					if (user && user.email) {
						const { device, ipAddress } =
							getRequestContextFromHeaders(ctx.request?.headers);
						console.log(
							`[Login] Sending login alert to ${user.email}`,
						);
						void sendLoginAlert(user.email as string, {
							userName: (user.name as string) || "",
							userEmail: user.email as string,
							device,
							ipAddress,
							loginTime: new Date(),
						});
					}
				}
			}

			// Send email notification when 2FA is verified/enabled
			// This happens after verify-otp (email) or verify-totp (app) - NOT on /enable
			if (
				ctx.path === "/two-factor/verify-otp" ||
				ctx.path === "/two-factor/verify-totp"
			) {
				const returned = ctx.context.returned as
					| Record<string, unknown>
					| undefined;
				// Check if verification was successful (returns user data)
				if (returned && "user" in returned) {
					const session = ctx.context.session;
					if (session?.user) {
						const user = session.user;
						console.log(
							`[2FA] Sending enable notification to ${user.email}`,
						);
						void sendTwoFactorEnabled(user.email, {
							userName: user.name || "",
							userEmail: user.email,
							enabledAt: new Date(),
						});
					}
				}
			}

			// Send email notification when 2FA is disabled
			if (ctx.path === "/two-factor/disable") {
				const returned = ctx.context.returned as
					| Record<string, unknown>
					| undefined;
				// Check if disable was successful
				if (returned && returned.status === true) {
					const session = ctx.context.session;
					if (session?.user) {
						const user = session.user;
						console.log(
							`[2FA] Sending disable notification to ${user.email}`,
						);
						void sendTwoFactorDisabled(user.email, {
							userName: user.name || "",
							userEmail: user.email,
							disabledAt: new Date(),
						});
					}
				}
			}
		}),
	},

	// Callbacks for invitation-only access
	databaseHooks: {
		user: {
			create: {
				before: async (user) => {
					// Check if email is authorized before creating user
					const { authorized, role } = await verifyEmailAuthorization(
						user.email,
					);

					if (!authorized) {
						throw new Error(
							"UNAUTHORIZED: Este email nao esta autorizado. Solicite um convite ao administrador.",
						);
					}

					// Set user role from invitation
					return {
						data: {
							...user,
							role: role,
						},
					};
				},
				after: async (user) => {
					// Mark invitation as accepted after user is created
					const { authorized, role } = await verifyEmailAuthorization(
						user.email,
					);

					if (authorized) {
						await markInvitationAccepted(user.id, user.email, role);
					}
				},
			},
		},
		session: {
			create: {
				before: async (session) => {
					// Check if user is banned
					const user = await prisma.user.findUnique({
						where: { id: session.userId },
					});

					if (user?.banned) {
						const banExpired =
							user.banExpiresAt && new Date() > user.banExpiresAt;
						if (!banExpired) {
							throw new Error(
								`Usuario banido: ${user.banReason || "Contate o administrador"}`,
							);
						}
					}

					return { data: session };
				},
			},
		},
	},
});

// Export auth types for use in components
export type Session = typeof auth.$Infer.Session;
export type User = typeof auth.$Infer.Session.user;
