import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { getActiveSubscription } from "@/server/services/subscription-service";
import { CardsSkeleton } from "@/components/shared/loading-skeleton";
import { PlansClient } from "./plans-client";

export const metadata: Metadata = {
	title: "Planos | Zé da API Manager",
};

async function PlansData({ userId }: { userId: string }) {
	const [plans, subscription] = await Promise.all([
		db.plan.findMany({
			where: { active: true },
			orderBy: { sortOrder: "asc" },
		}),
		getActiveSubscription(userId),
	]);

	return (
		<PlansClient
			plans={plans.map((p) => ({
				...p,
				price: Number(p.price),
				features: p.features as string[],
			}))}
			currentPlanSlug={subscription?.plan.slug ?? null}
		/>
	);
}

export default async function PlansPage() {
	const session = await requireAuth();

	return (
		<div className="mx-auto max-w-6xl space-y-6 py-4">
			<div className="text-center">
				<h1 className="text-3xl font-medium tracking-tighter sm:text-4xl">
					Pague conforme escalar
				</h1>
				<p className="mx-auto mt-2 max-w-xl text-sm text-muted-foreground">
					Escolha o numero de instancias WhatsApp. Ajuste quando
					quiser. Sem compromissos.
				</p>
			</div>

			<Suspense fallback={<CardsSkeleton count={1} />}>
				<PlansData userId={session.user.id} />
			</Suspense>
		</div>
	);
}
