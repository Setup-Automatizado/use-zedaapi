"use client";

import { useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Building2, Loader2 } from "lucide-react";

interface OrganizationFormProps {
	user: {
		name: string;
		email: string;
	};
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function OrganizationForm({ user }: OrganizationFormProps) {
	const [saving, setSaving] = useState(false);

	const handleSave = async (e: React.FormEvent) => {
		e.preventDefault();
		setSaving(true);
		// TODO: Call server action to update organization
		await new Promise((r) => setTimeout(r, 1000));
		setSaving(false);
	};

	return (
		<form onSubmit={handleSave}>
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Building2 className="size-4" />
						Dados da Organizacao
					</CardTitle>
					<CardDescription>
						Informacoes basicas da sua organizacao.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="grid gap-4 sm:grid-cols-2">
						<div className="space-y-2">
							<Label htmlFor="org-name">
								Nome da Organizacao
							</Label>
							<Input id="org-name" placeholder="Minha Empresa" />
						</div>
						<div className="space-y-2">
							<Label htmlFor="org-slug">Slug</Label>
							<Input id="org-slug" placeholder="minha-empresa" />
							<p className="text-xs text-muted-foreground">
								Identificador unico da organizacao.
							</p>
						</div>
					</div>

					<div className="space-y-2">
						<Label>Logo</Label>
						<div className="flex items-center gap-4">
							<div className="flex size-16 items-center justify-center rounded-xl bg-muted">
								<Building2 className="size-6 text-muted-foreground" />
							</div>
							<div>
								<Button
									type="button"
									variant="outline"
									size="sm"
								>
									Enviar logo
								</Button>
								<p className="mt-1 text-xs text-muted-foreground">
									PNG ou JPG, max 2MB
								</p>
							</div>
						</div>
					</div>

					<div className="flex justify-end pt-2">
						<Button type="submit" disabled={saving}>
							{saving && (
								<Loader2 className="size-4 animate-spin" />
							)}
							Salvar alteracoes
						</Button>
					</div>
				</CardContent>
			</Card>
		</form>
	);
}
