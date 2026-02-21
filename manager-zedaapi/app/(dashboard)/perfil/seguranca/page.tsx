import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { PageHeader } from "@/components/shared/page-header";
import { SecurityClient } from "@/components/profile/security-client";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Segurança | Zé da API Manager",
};

export default async function SecurityPage() {
	await requireAuth();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Segurança"
				description="Gerencie a segurança da sua conta."
			/>

			<div className="flex gap-2 text-sm">
				<Link
					href="/perfil"
					className="rounded-lg px-3 py-1.5 text-muted-foreground hover:bg-muted/50"
				>
					Geral
				</Link>
				<Link
					href="/perfil/seguranca"
					className="rounded-lg bg-muted px-3 py-1.5 font-medium"
				>
					Segurança
				</Link>
			</div>

			<SecurityClient />
		</div>
	);
}
