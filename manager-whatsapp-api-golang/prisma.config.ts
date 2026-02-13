import { defineConfig } from "prisma/config";

// pg v8+ treats sslmode=require as verify-full.
// RDS AWS uses Amazon CA, so use no-verify for migrations.
const url = (process.env["DATABASE_URL"] ?? "").replace(
	"sslmode=require",
	"sslmode=no-verify",
);

export default defineConfig({
	schema: "prisma/schema.prisma",
	migrations: {
		path: "prisma/migrations",
		seed: "bun prisma/seed.ts",
	},
	datasource: { url },
});
