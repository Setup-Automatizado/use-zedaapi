import { PrismaPg } from "@prisma/adapter-pg";
import { PrismaClient } from "@/lib/generated/prisma/client";

const globalForPrisma = global as unknown as {
	prisma: PrismaClient | undefined;
};

function createPrismaClient(): PrismaClient {
	const adapter = new PrismaPg({
		connectionString: process.env.DATABASE_URL!,
		ssl:
			process.env.NODE_ENV === "production"
				? { rejectUnauthorized: false }
				: undefined,
	});
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
