import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { SettingsClient } from "./settings-client";

export const metadata: Metadata = {
	title: "Configuracoes | Zé da API Manager",
};

export default async function SettingsPage() {
	await requireAuth();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Configuracoes
				</h1>
				<p className="text-sm text-muted-foreground">
					Gerencie as preferencias da sua conta.
				</p>
			</div>

			<SettingsClient />
		</div>
	);
}
