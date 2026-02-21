import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getAdminPlans } from "@/server/actions/admin";
import { CardsSkeleton } from "@/components/shared/loading-skeleton";
import { PlansContentClient } from "./plans-content-client";

export const metadata: Metadata = {
	title: "Planos | Admin Zé da API Manager",
};

export default async function AdminPlansPage() {
	await requireAdmin();

	return (
		<Suspense fallback={<PlansPageSkeleton />}>
			<PlansContent />
		</Suspense>
	);
}

async function PlansContent() {
	const res = await getAdminPlans();
	const plans = res.data ?? [];

	return <PlansContentClient initialPlans={plans} />;
}

function PlansPageSkeleton() {
	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<div className="h-7 w-32 animate-pulse rounded bg-muted" />
					<div className="mt-1 h-4 w-56 animate-pulse rounded bg-muted" />
				</div>
				<div className="h-9 w-28 animate-pulse rounded bg-muted" />
			</div>
			<CardsSkeleton count={3} />
		</div>
	);
}
