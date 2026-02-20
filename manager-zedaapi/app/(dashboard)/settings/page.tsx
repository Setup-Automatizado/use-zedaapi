import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { PageHeader } from "@/components/shared/page-header";
import { SettingsClient } from "./settings-client";

export const metadata: Metadata = {
	title: "Configurações | Zé da API Manager",
};

export default async function SettingsPage() {
	await requireAuth();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Configurações"
				description="Gerencie as preferências da sua conta."
			/>

			<SettingsClient />
		</div>
	);
}
