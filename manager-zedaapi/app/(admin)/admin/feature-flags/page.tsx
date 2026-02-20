import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getFeatureFlags } from "@/server/actions/admin";
import { CardsSkeleton } from "@/components/shared/loading-skeleton";
import { FeatureFlagsClient } from "./feature-flags-client";

export const metadata: Metadata = {
	title: "Feature Flags | Admin Zé da API Manager",
};

export default async function AdminFeatureFlagsPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Feature Flags
				</h1>
				<p className="text-sm text-muted-foreground">
					Controle funcionalidades da plataforma.
				</p>
			</div>

			<Suspense fallback={<CardsSkeleton count={3} />}>
				<FeatureFlagsContent />
			</Suspense>
		</div>
	);
}

async function FeatureFlagsContent() {
	const res = await getFeatureFlags();
	const flags = res.data ?? [];

	return <FeatureFlagsClient initialFlags={flags} />;
}
