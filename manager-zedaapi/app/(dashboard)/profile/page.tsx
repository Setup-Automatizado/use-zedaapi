import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { ProfileFormSkeleton } from "@/components/shared/loading-skeleton";
import { ProfileForm } from "@/components/profile/profile-form";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Perfil | Zé da API Manager",
};

async function ProfileData() {
	const session = await requireAuth();

	return (
		<ProfileForm
			user={{
				name: session.user.name ?? "",
				email: session.user.email ?? "",
			}}
		/>
	);
}

export default async function ProfilePage() {
	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Perfil</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie suas informacoes pessoais.
				</p>
			</div>

			<div className="flex gap-2 text-sm">
				<Link
					href="/profile"
					className="rounded-lg bg-muted px-3 py-1.5 font-medium"
				>
					Geral
				</Link>
				<Link
					href="/profile/security"
					className="rounded-lg px-3 py-1.5 text-muted-foreground hover:bg-muted/50"
				>
					Seguranca
				</Link>
			</div>

			<Suspense fallback={<ProfileFormSkeleton />}>
				<ProfileData />
			</Suspense>
		</div>
	);
}
