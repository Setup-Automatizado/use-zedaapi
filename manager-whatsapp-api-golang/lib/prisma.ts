import { PrismaPg } from "@prisma/adapter-pg";
import { PrismaClient } from "@/lib/generated/prisma/client";

const globalForPrisma = global as unknown as {
	prisma: PrismaClient | undefined;
};

function createPrismaClient(): PrismaClient {
	let connectionString = process.env.DATABASE_URL!;
	// pg v8+ treats sslmode=require as verify-full (validates certificate).
	// RDS AWS uses Amazon CA (self-signed chain), so use no-verify instead.
	if (process.env.NODE_ENV === "production") {
		connectionString = connectionString.replace(
			"sslmode=require",
			"sslmode=no-verify",
		);
	}
	const adapter = new PrismaPg({ connectionString });
	return new PrismaClient({ adapter });
}

function getPrismaClient(): PrismaClient {
	if (!globalForPrisma.prisma) {
		globalForPrisma.prisma = createPrismaClient();
	}
	return globalForPrisma.prisma;
}

const prisma = new Proxy({} as PrismaClient, {
	get(_target, prop) {
		const client = getPrismaClient();
		const value = Reflect.get(client, prop);
		if (typeof value === "function") {
			return value.bind(client);
		}
		return value;
	},
});

export default prisma;
