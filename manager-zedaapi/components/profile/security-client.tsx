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
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
	Loader2,
	KeyRound,
	ShieldCheck,
	Monitor,
	Smartphone,
} from "lucide-react";

export function SecurityClient() {
	const [changingPassword, setChangingPassword] = useState(false);
	const [twoFactorEnabled] = useState(false);

	const mockSessions = [
		{
			id: "1",
			device: "Chrome no macOS",
			icon: Monitor,
			ip: "189.40.xxx.xxx",
			lastActive: "Agora",
			current: true,
		},
		{
			id: "2",
			device: "Safari no iPhone",
			icon: Smartphone,
			ip: "189.40.xxx.xxx",
			lastActive: "2 horas atras",
			current: false,
		},
	];

	const handleChangePassword = async (e: React.FormEvent) => {
		e.preventDefault();
		setChangingPassword(true);
		await new Promise((r) => setTimeout(r, 1000));
		setChangingPassword(false);
	};

	return (
		<>
			<form onSubmit={handleChangePassword}>
				<Card>
					<CardHeader>
						<CardTitle className="flex items-center gap-2">
							<KeyRound className="size-4" />
							Alterar Senha
						</CardTitle>
						<CardDescription>
							Escolha uma senha forte com pelo menos 8 caracteres.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="max-w-sm space-y-2">
							<Label htmlFor="current-password">
								Senha atual
							</Label>
							<Input
								id="current-password"
								type="password"
								placeholder="Sua senha atual"
							/>
						</div>
						<div className="max-w-sm space-y-2">
							<Label htmlFor="new-password">Nova senha</Label>
							<Input
								id="new-password"
								type="password"
								placeholder="Nova senha"
							/>
						</div>
						<div className="max-w-sm space-y-2">
							<Label htmlFor="confirm-password">
								Confirmar nova senha
							</Label>
							<Input
								id="confirm-password"
								type="password"
								placeholder="Confirmar nova senha"
							/>
						</div>
						<Button type="submit" disabled={changingPassword}>
							{changingPassword && (
								<Loader2 className="size-4 animate-spin" />
							)}
							Alterar senha
						</Button>
					</CardContent>
				</Card>
			</form>

			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<ShieldCheck className="size-4" />
						Autenticacao em Dois Fatores (2FA)
					</CardTitle>
					<CardDescription>
						Adicione uma camada extra de seguranca a sua conta.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex items-center justify-between">
						<div className="flex items-center gap-3">
							<Badge
								variant={
									twoFactorEnabled ? "default" : "secondary"
								}
							>
								{twoFactorEnabled ? "Ativado" : "Desativado"}
							</Badge>
							<span className="text-sm text-muted-foreground">
								{twoFactorEnabled
									? "Sua conta esta protegida com 2FA."
									: "Recomendamos ativar para maior seguranca."}
							</span>
						</div>
						<Button
							variant={
								twoFactorEnabled ? "destructive" : "default"
							}
							size="sm"
						>
							{twoFactorEnabled
								? "Desativar 2FA"
								: "Ativar 2FA"}
						</Button>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Monitor className="size-4" />
						Sessoes Ativas
					</CardTitle>
					<CardDescription>
						Dispositivos com acesso a sua conta.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-3">
					{mockSessions.map((session, i) => (
						<div key={session.id}>
							{i > 0 && <Separator className="mb-3" />}
							<div className="flex items-center justify-between">
								<div className="flex items-center gap-3">
									<div className="flex size-9 items-center justify-center rounded-lg bg-muted">
										<session.icon className="size-4 text-muted-foreground" />
									</div>
									<div>
										<p className="text-sm font-medium">
											{session.device}
											{session.current && (
												<Badge
													variant="secondary"
													className="ml-2"
												>
													Atual
												</Badge>
											)}
										</p>
										<p className="text-xs text-muted-foreground">
											{session.ip} -{" "}
											{session.lastActive}
										</p>
									</div>
								</div>
								{!session.current && (
									<Button
										variant="ghost"
										size="sm"
										className="text-destructive"
									>
										Encerrar
									</Button>
								)}
							</div>
						</div>
					))}
				</CardContent>
			</Card>
		</>
	);
}
