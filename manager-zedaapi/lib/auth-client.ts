"use client";

import { createAuthClient } from "better-auth/react";
import { adminClient } from "better-auth/client/plugins";
import { organizationClient } from "better-auth/client/plugins";
import { twoFactorClient } from "better-auth/client/plugins";
import { waitlistClient } from "@guilhermejansen/better-auth-waitlist/client";

export const authClient = createAuthClient({
	baseURL: process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000",
	plugins: [
		adminClient(),
		organizationClient(),
		twoFactorClient(),
		waitlistClient(),
	],
});

export const { signIn, signUp, signOut, useSession, getSession } = authClient;
