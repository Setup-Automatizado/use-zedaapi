import { notFound } from "next/navigation";
import { requireAuth } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { InstanceDetailClient } from "@/components/instances/instance-detail-client";

type PageProps = {
	params: Promise<{ id: string }>;
};

export async function generateMetadata({ params }: PageProps) {
	const { id } = await params;
	const instance = await db.instance.findUnique({
		where: { id },
		select: { name: true },
	});

	return {
		title: instance
			? `${instance.name} | Zé da API Manager`
			: "Instancia | Zé da API Manager",
	};
}

export default async function InstanceDetailPage({ params }: PageProps) {
	const session = await requireAuth();
	const { id } = await params;

	const instance = await db.instance.findUnique({
		where: { id, userId: session.user.id },
	});

	if (!instance) {
		notFound();
	}

	return (
		<InstanceDetailClient
			instance={{
				id: instance.id,
				name: instance.name,
				status: instance.status,
				phone: instance.phone,
				lastSyncAt: instance.lastSyncAt?.toISOString() ?? null,
			}}
		/>
	);
}
