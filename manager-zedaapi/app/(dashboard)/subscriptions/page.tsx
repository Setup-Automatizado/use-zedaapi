import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { getActiveSubscription } from "@/server/services/subscription-service";
import { CardSkeleton } from "@/components/shared/loading-skeleton";
import { SubscriptionDetails } from "./subscription-details";

export const metadata: Metadata = {
	title: "Assinatura | Zé da API Manager",
};

async function SubscriptionData({ userId }: { userId: string }) {
	const subscription = await getActiveSubscription(userId);
	return <SubscriptionDetails subscription={subscription} />;
}

export default async function SubscriptionsPage() {
	const session = await requireAuth();

	return (
		<div className="mx-auto max-w-4xl space-y-8">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Assinatura
				</h1>
				<p className="text-sm text-muted-foreground mt-1">
					Gerencie seu plano e assinatura
				</p>
			</div>

			<Suspense fallback={<CardSkeleton />}>
				<SubscriptionData userId={session.user.id} />
			</Suspense>
		</div>
	);
}
