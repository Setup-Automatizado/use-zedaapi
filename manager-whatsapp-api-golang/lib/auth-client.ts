/**
 * Better Auth Client
 *
 * Client-side authentication hooks and functions.
 * Includes two-factor authentication and admin support.
 *
 * @module lib/auth-client
 */

import { createAuthClient } from "better-auth/react";
import { twoFactorClient, adminClient } from "better-auth/client/plugins";
import { ac, adminRole, userRole } from "@/lib/auth/permissions";

const authClient = createAuthClient({
	baseURL: process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000",
	plugins: [
		twoFactorClient({
			/**
			 * Redirect to 2FA verification page when required
			 * This is triggered after successful password authentication
			 * when the user has 2FA enabled
			 */
			onTwoFactorRedirect() {
				window.location.href = "/verify-2fa";
			},
		}),
		adminClient({
			ac,
			roles: {
				ADMIN: adminRole,
				USER: userRole,
			},
		}),
	],
});

// Export all auth client methods
export const {
	signIn,
	signUp,
	signOut,
	useSession,
	// Two-factor authentication
	twoFactor,
	// Admin functions
	admin,
} = authClient;

// Re-export the client for advanced usage
export default authClient;
