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
		requireEmailVerification: false,

		sendResetPassword: async ({
			user,
			url,
		}: {
			user: { email: string; name?: string | null };
			url: string;
		}) => {
			const { sendTemplateEmail } = await import("@/lib/email");
			await sendTemplateEmail(user.email, "magic-link", {
				userName: user.name || "Usuário",
				resetUrl: url,
				subject: "Redefinição de senha — Zé da API Manager",
			});
		},
		sendVerificationEmail: async ({
			user,
			url,
		}: {
			user: { email: string; name?: string | null };
			url: string;
		}) => {
			const { sendTemplateEmail } = await import("@/lib/email");
			await sendTemplateEmail(user.email, "magic-link", {
				userName: user.name || "Usuário",
				verifyUrl: url,
				subject: "Verifique seu e-mail — Zé da API Manager",
			});
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
			sendInviteEmail: async ({ email, inviteCode }) => {
				const { sendTemplateEmail } = await import("@/lib/email");
				const signUpUrl = `${process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000"}/cadastro`;
				await sendTemplateEmail(email, "waitlist-approved", {
					userName: email.split("@")[0] || "Usuário",
					inviteCode,
					signUpUrl,
					subject: "Sua conta foi aprovada! — Zé da API Manager",
				});
			},
		}),
		nextCookies(),
	],
});

export type Session = typeof auth.$Infer.Session;
export type User = Session["user"];
