import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { InstancesClient } from "@/components/instances/instances-client";

export const metadata: Metadata = {
	title: "Instancias | Zé da API Manager",
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
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-bold tracking-tight">
						Instancias
					</h1>
					<p className="text-sm text-muted-foreground">
						Gerencie suas instancias WhatsApp.
					</p>
				</div>
			</div>

			<Suspense fallback={<TableSkeleton />}>
				<InstancesList userId={session.user.id} />
			</Suspense>
		</div>
	);
}
