import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { getSystemSettings } from "@/server/actions/admin";
import { Card, CardContent } from "@/components/ui/card";
import { FormSkeleton } from "@/components/shared/loading-skeleton";
import { SettingsFormClient } from "./settings-form-client";

export const metadata: Metadata = {
	title: "Configuracoes | Admin Zé da API Manager",
};

export default async function AdminSettingsPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">
					Configuracoes do Sistema
				</h1>
				<p className="text-sm text-muted-foreground">
					Configuracoes globais da plataforma.
				</p>
			</div>

			<Suspense
				fallback={
					<Card>
						<CardContent className="p-6">
							<FormSkeleton fields={5} />
						</CardContent>
					</Card>
				}
			>
				<SettingsContent />
			</Suspense>
		</div>
	);
}

async function SettingsContent() {
	const res = await getSystemSettings();
	const settings = res.data ?? [];

	if (settings.length === 0) {
		return (
			<Card>
				<CardContent className="py-12 text-center">
					<p className="text-sm text-muted-foreground">
						Nenhuma configuracao encontrada.
					</p>
				</CardContent>
			</Card>
		);
	}

	return <SettingsFormClient initialSettings={settings} />;
}
