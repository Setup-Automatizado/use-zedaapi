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
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { authClient } from "@/lib/auth-client";

interface ProfileFormProps {
	user: {
		name: string;
		email: string;
	};
}

export function ProfileForm({ user }: ProfileFormProps) {
	const [name, setName] = useState(user.name);
	const [saving, setSaving] = useState(false);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSaving(true);
		try {
			await authClient.updateUser({ name });
			toast.success("Perfil atualizado");
		} catch {
			toast.error("Erro ao atualizar perfil");
		} finally {
			setSaving(false);
		}
	}

	return (
		<form onSubmit={handleSubmit}>
			<Card>
				<CardHeader>
					<CardTitle>Informacoes Pessoais</CardTitle>
					<CardDescription>
						Atualize suas informacoes de perfil.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="space-y-2">
						<Label htmlFor="name">Nome</Label>
						<Input
							id="name"
							value={name}
							onChange={(e) => setName(e.target.value)}
							placeholder="Seu nome"
						/>
					</div>

					<div className="space-y-2">
						<Label htmlFor="email">Email</Label>
						<Input
							id="email"
							value={user.email}
							disabled
							className="bg-muted"
						/>
						<p className="text-xs text-muted-foreground">
							O email nao pode ser alterado.
						</p>
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
