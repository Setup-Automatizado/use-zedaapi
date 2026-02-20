import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { Suspense } from "react";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { ApiKeysClient } from "./api-keys-client";

export const metadata: Metadata = {
	title: "Chaves API | Zé da API Manager",
};

async function ApiKeysList({ userId }: { userId: string }) {
	const keys = await db.apiKey.findMany({
		where: { userId, active: true },
		select: {
			id: true,
			name: true,
			key: true,
			createdAt: true,
			lastUsedAt: true,
		},
		orderBy: { createdAt: "desc" },
	});

	return (
		<ApiKeysClient
			keys={keys.map((k) => ({
				id: k.id,
				name: k.name,
				prefix: `${k.key.slice(0, 8)}...${k.key.slice(-4)}`,
				createdAt: k.createdAt.toISOString(),
				lastUsed: k.lastUsedAt?.toISOString() ?? null,
			}))}
		/>
	);
}

export default async function ApiKeysPage() {
	const session = await requireAuth();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Chaves API
				</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie as chaves de acesso a API.
				</p>
			</div>

			<Suspense fallback={<TableSkeleton rows={3} />}>
				<ApiKeysList userId={session.user.id} />
			</Suspense>
		</div>
	);
}
