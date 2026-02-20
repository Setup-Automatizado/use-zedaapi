"use client";

import { useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Bell, Globe, Trash2 } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";

export function SettingsClient() {
	const [deleteAccountDialogOpen, setDeleteAccountDialogOpen] =
		useState(false);

	return (
		<>
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Bell className="size-4" />
						Notificacoes
					</CardTitle>
					<CardDescription>
						Escolha como deseja receber notificacoes.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="flex items-center justify-between">
						<div className="space-y-0.5">
							<Label>Email de status</Label>
							<p className="text-xs text-muted-foreground">
								Receba alertas quando uma instancia desconectar.
							</p>
						</div>
						<Switch defaultChecked />
					</div>
					<Separator />
					<div className="flex items-center justify-between">
						<div className="space-y-0.5">
							<Label>Resumo semanal</Label>
							<p className="text-xs text-muted-foreground">
								Receba um resumo semanal de uso e metricas.
							</p>
						</div>
						<Switch />
					</div>
					<Separator />
					<div className="flex items-center justify-between">
						<div className="space-y-0.5">
							<Label>Alertas de faturamento</Label>
							<p className="text-xs text-muted-foreground">
								Notificacoes sobre faturas e pagamentos.
							</p>
						</div>
						<Switch defaultChecked />
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Globe className="size-4" />
						Idioma e Regiao
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="flex items-center justify-between">
						<div className="space-y-0.5">
							<Label>Idioma</Label>
							<p className="text-xs text-muted-foreground">
								Idioma da interface.
							</p>
						</div>
						<Badge variant="outline">Portugues (BR)</Badge>
					</div>
				</CardContent>
			</Card>

			<Card className="border-destructive/20">
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-destructive">
						<Trash2 className="size-4" />
						Zona de Perigo
					</CardTitle>
					<CardDescription>
						Acoes irreversiveis. Prossiga com cautela.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex items-center justify-between">
						<div>
							<p className="text-sm font-medium">Excluir conta</p>
							<p className="text-xs text-muted-foreground">
								Exclua permanentemente sua conta e todos os
								dados associados.
							</p>
						</div>
						<Button
							variant="destructive"
							size="sm"
							onClick={() => setDeleteAccountDialogOpen(true)}
						>
							Excluir Conta
						</Button>
					</div>
				</CardContent>
			</Card>

			<ConfirmDialog
				open={deleteAccountDialogOpen}
				onOpenChange={setDeleteAccountDialogOpen}
				title="Excluir conta"
				description="Esta acao e irreversivel. Todos os seus dados, instancias e configuracoes serao permanentemente removidos. Digite 'excluir' para confirmar."
				confirmLabel="Excluir conta permanentemente"
				destructive
				onConfirm={() => setDeleteAccountDialogOpen(false)}
			/>
		</>
	);
}
