import { waitlist } from "@guilhermejansen/better-auth-waitlist";
import { betterAuth } from "better-auth";
import { prismaAdapter } from "better-auth/adapters/prisma";
import { nextCookies } from "better-auth/next-js";
import { admin } from "better-auth/plugins/admin";
import { organization } from "better-auth/plugins/organization";
import { twoFactor } from "better-auth/plugins/two-factor";
import { db } from "@/lib/db";

export const auth = betterAuth({
	database: prismaAdapter(db, {
		provider: "postgresql",
	}),
	baseURL: process.env.BETTER_AUTH_URL,
	secret: process.env.BETTER_AUTH_SECRET,
	trustedOrigins: process.env.BETTER_AUTH_TRUSTED_ORIGINS?.split(",") ?? [],

	emailAndPassword: {
		enabled: true,
		requireEmailVerification: true,
		sendResetPassword: async ({
			user,
			url,
		}: {
			user: { email: string };
			url: string;
		}) => {
			// TODO: Send via SMTP email service
			// Do NOT log sensitive URLs
		},
		sendVerificationEmail: async ({
			user,
			url,
		}: {
			user: { email: string };
			url: string;
		}) => {
			// TODO: Send via SMTP email service
			// Do NOT log sensitive URLs
		},
	},

	session: {
		expiresIn: 60 * 60 * 24 * 7, // 7 days
		updateAge: 60 * 60 * 24, // 1 day
		cookieCache: {
			enabled: true,
			maxAge: 60 * 5, // 5 minutes
		},
	},

	account: {
		accountLinking: {
			enabled: true,
		},
	},

	plugins: [
		admin(),
		organization(),
		twoFactor({
			issuer: "Zé da API Manager",
		}),
		waitlist({
			enabled: process.env.WAITLIST_ENABLED !== "false",
			requireInviteCode:
				process.env.WAITLIST_REQUIRE_APPROVAL !== "false",
		}),
		nextCookies(),
	],
});

export type Session = typeof auth.$Infer.Session;
export type User = Session["user"];
