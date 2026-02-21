import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { Suspense } from "react";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
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
			<PageHeader
				title="Chaves API"
				description="Gerencie as chaves de acesso à API."
			/>

			<Suspense fallback={<TableSkeleton rows={3} />}>
				<ApiKeysList userId={session.user.id} />
			</Suspense>
		</div>
	);
}
