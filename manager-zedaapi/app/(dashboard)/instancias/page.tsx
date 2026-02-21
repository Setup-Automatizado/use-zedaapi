import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";
import { InstancesClient } from "@/components/instances/instances-client";

export const metadata: Metadata = {
	title: "Instâncias | Zé da API Manager",
};

async function InstancesList({ userId }: { userId: string }) {
	const instances = await db.instance.findMany({
		where: { userId },
		orderBy: { createdAt: "desc" },
	});

	return <InstancesClient instances={instances} />;
}

export default async function InstancesPage() {
	const session = await requireAuth();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Instâncias"
				description="Gerencie suas instâncias WhatsApp."
			/>

			<Suspense fallback={<TableSkeleton />}>
				<InstancesList userId={session.user.id} />
			</Suspense>
		</div>
	);
}
