/**
 * Database Seed Script
 *
 * Creates the initial admin user for the WhatsApp Manager platform.
 * Uses the same password hashing as Better Auth for compatibility.
 *
 * Run with: bunx prisma db seed
 *
 * @module prisma/seed
 */

import { PrismaPg } from "@prisma/adapter-pg";
import { hashPassword as betterAuthHashPassword } from "better-auth/crypto";
import { PrismaClient } from "../lib/generated/prisma/client";

// Create Prisma client with PostgreSQL adapter
const adapter = new PrismaPg({
	connectionString: process.env.DATABASE_URL!,
});
const prisma = new PrismaClient({ adapter });

// Admin credentials
const ADMIN_EMAIL = "guilhermejansenoficial@gmail.com";
const ADMIN_PASSWORD = "Admin@2026@Manager#";
const ADMIN_NAME = "Admin";

/**
 * Hash password using Better Auth's internal function
 */
async function hashPassword(password: string): Promise<string> {
	return betterAuthHashPassword(password);
}

async function main() {
	console.log("Starting database seed...");

	// Check if admin already exists
	const existingUser = await prisma.user.findUnique({
		where: { email: ADMIN_EMAIL },
	});

	if (existingUser) {
		console.log("Admin user already exists.");

		// Update role and twoFactorEnabled
		await prisma.user.update({
			where: { email: ADMIN_EMAIL },
			data: {
				role: "ADMIN",
				twoFactorEnabled: false,
			},
		});
		console.log("Updated user: role=ADMIN, twoFactorEnabled=false");

		// Check if credential account exists
		const credentialAccount = await prisma.account.findFirst({
			where: {
				userId: existingUser.id,
				providerId: "credential",
			},
		});

		if (!credentialAccount) {
			console.log("Creating credential account for existing user...");
			const hashedPassword = await hashPassword(ADMIN_PASSWORD);

			await prisma.account.create({
				data: {
					id: crypto.randomUUID(),
					userId: existingUser.id,
					accountId: existingUser.id,
					providerId: "credential",
					password: hashedPassword,
				},
			});
			console.log("Credential account created!");
		} else {
			// Update password
			console.log("Updating credential account password...");
			const hashedPassword = await hashPassword(ADMIN_PASSWORD);

			await prisma.account.update({
				where: { id: credentialAccount.id },
				data: { password: hashedPassword },
			});
			console.log("Password updated!");
		}

		// Ensure AllowedUser entry exists
		await prisma.allowedUser.upsert({
			where: { email: ADMIN_EMAIL },
			update: {
				role: "ADMIN",
				acceptedAt: new Date(),
				userId: existingUser.id,
			},
			create: {
				email: ADMIN_EMAIL,
				role: "ADMIN",
				acceptedAt: new Date(),
				userId: existingUser.id,
			},
		});
		console.log("AllowedUser entry ensured.");

		console.log("\nAdmin user updated successfully!");
		return;
	}

	// Create new admin user
	const userId = crypto.randomUUID();
	const hashedPassword = await hashPassword(ADMIN_PASSWORD);

	console.log("Creating new admin user...");

	// Create user first
	await prisma.user.create({
		data: {
			id: userId,
			email: ADMIN_EMAIL,
			name: ADMIN_NAME,
			role: "ADMIN",
			emailVerified: true,
			twoFactorEnabled: false,
		},
	});

	// Create AllowedUser entry (userId now exists)
	await prisma.allowedUser.upsert({
		where: { email: ADMIN_EMAIL },
		update: {
			role: "ADMIN",
			acceptedAt: new Date(),
			userId: userId,
		},
		create: {
			email: ADMIN_EMAIL,
			role: "ADMIN",
			acceptedAt: new Date(),
			userId: userId,
		},
	});

	// Create credential account
	await prisma.account.create({
		data: {
			id: crypto.randomUUID(),
			userId: userId,
			accountId: userId,
			providerId: "credential",
			password: hashedPassword,
		},
	});

	console.log("");
	console.log("=".repeat(50));
	console.log("Admin user created successfully!");
	console.log("=".repeat(50));
	console.log(`Email: ${ADMIN_EMAIL}`);
	console.log(`Password: ${ADMIN_PASSWORD}`);
	console.log(`Role: ADMIN`);
	console.log("=".repeat(50));
	console.log("");
}

main()
	.catch((e) => {
		console.error("Seed failed:", e);
		process.exit(1);
	})
	.finally(async () => {
		await prisma.$disconnect();
	});
