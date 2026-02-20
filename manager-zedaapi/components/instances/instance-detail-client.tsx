"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { cn } from "@/lib/utils";
import {
	ArrowLeft,
	Copy,
	PowerOff,
	QrCode,
	RefreshCw,
	Trash2,
} from "lucide-react";
import { toast } from "sonner";
import Link from "next/link";

interface InstanceData {
	id: string;
	name: string;
	status: string;
	phone: string | null;
	lastSyncAt: string | null;
}

interface InstanceDetailClientProps {
	instance: InstanceData;
}

const statusConfig: Record<
	string,
	{ label: string; className: string; dot: string }
> = {
	connected: {
		label: "Conectado",
		className: "bg-primary/10 text-primary",
		dot: "bg-primary",
	},
	connecting: {
		label: "Conectando",
		className: "bg-chart-2/10 text-chart-2",
		dot: "bg-chart-2",
	},
	disconnected: {
		label: "Desconectado",
		className: "bg-muted text-muted-foreground",
		dot: "bg-muted-foreground",
	},
	creating: {
		label: "Criando",
		className: "bg-chart-2/10 text-chart-2",
		dot: "bg-chart-2",
	},
	error: {
		label: "Erro",
		className: "bg-destructive/10 text-destructive",
		dot: "bg-destructive",
	},
	banned: {
		label: "Banido",
		className: "bg-destructive/10 text-destructive",
		dot: "bg-destructive",
	},
};

export function InstanceDetailClient({ instance }: InstanceDetailClientProps) {
	const router = useRouter();
	const [deleteDialog, setDeleteDialog] = useState(false);
	const [deleteLoading, setDeleteLoading] = useState(false);

	const config = statusConfig[instance.status] ?? {
		label: instance.status,
		className: "bg-muted text-muted-foreground",
		dot: "bg-muted-foreground",
	};

	const isConnected = instance.status === "connected";

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
		toast.success("Copiado para a area de transferencia");
	}

	async function handleDelete() {
		setDeleteLoading(true);
		try {
			// TODO: Wire to real delete action
			toast.success("Instancia excluida");
			router.push("/dashboard/instances");
		} catch {
			toast.error("Erro ao excluir instancia");
		} finally {
			setDeleteLoading(false);
			setDeleteDialog(false);
		}
	}

	return (
		<div className="space-y-6">
			<div className="flex items-center gap-4">
				<Button variant="ghost" size="icon-sm" asChild>
					<Link href="/dashboard/instances">
						<ArrowLeft className="size-4" />
					</Link>
				</Button>
				<div className="flex-1">
					<div className="flex items-center gap-3">
						<h1 className="text-2xl font-bold tracking-tight">
							{instance.name}
						</h1>
						<Badge
							variant="secondary"
							className={cn(config.className)}
						>
							<span
								className={cn(
									"mr-1.5 inline-block size-2 rounded-full",
									config.dot,
								)}
							/>
							{config.label}
						</Badge>
					</div>
					<p className="text-sm text-muted-foreground">
						ID: {instance.id}
					</p>
				</div>
				<div className="flex items-center gap-2">
					{isConnected ? (
						<Button variant="outline" size="sm">
							<PowerOff className="mr-2 size-4" />
							Desconectar
						</Button>
					) : (
						<Button variant="outline" size="sm">
							<QrCode className="mr-2 size-4" />
							Conectar
						</Button>
					)}
					<Button variant="outline" size="sm">
						<RefreshCw className="mr-2 size-4" />
						Reiniciar
					</Button>
				</div>
			</div>

			<div className="grid gap-6 md:grid-cols-2">
				<Card>
					<CardHeader>
						<CardTitle className="text-base">Informacoes</CardTitle>
						<CardDescription>
							Detalhes da instancia WhatsApp
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="space-y-2">
							<Label className="text-muted-foreground">
								Telefone
							</Label>
							<div className="flex items-center gap-2">
								<p className="text-sm font-medium">
									{instance.phone ?? "Nao conectado"}
								</p>
								{instance.phone && (
									<Button
										variant="ghost"
										size="icon-sm"
										onClick={() =>
											copyToClipboard(instance.phone!)
										}
									>
										<Copy className="size-3" />
									</Button>
								)}
							</div>
						</div>

						<Separator />

						<div className="space-y-2">
							<Label className="text-muted-foreground">
								Ultimo sync
							</Label>
							<p className="text-sm font-medium">
								{instance.lastSyncAt
									? new Date(
											instance.lastSyncAt,
										).toLocaleString("pt-BR")
									: "-"}
							</p>
						</div>

						<Separator />

						<div className="space-y-2">
							<Label className="text-muted-foreground">
								ID da Instancia
							</Label>
							<div className="flex items-center gap-2">
								<code className="text-xs bg-muted px-2 py-1 rounded-md font-mono">
									{instance.id}
								</code>
								<Button
									variant="ghost"
									size="icon-sm"
									onClick={() => copyToClipboard(instance.id)}
								>
									<Copy className="size-3" />
								</Button>
							</div>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">Status</CardTitle>
						<CardDescription>
							Status da conexao WhatsApp
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="flex items-center gap-3">
							<span
								className={cn(
									"inline-block size-3 rounded-full",
									config.dot,
								)}
							/>
							<span className="text-sm font-medium">
								{config.label}
							</span>
						</div>
						<p className="text-xs text-muted-foreground">
							A configuracao de webhooks e feita diretamente na
							API Zé da API.
						</p>
					</CardContent>
				</Card>
			</div>

			<Card className="border-destructive/50">
				<CardHeader>
					<CardTitle className="text-base text-destructive">
						Zona de Perigo
					</CardTitle>
					<CardDescription>
						Acoes irreversiveis para esta instancia.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<Button
						variant="destructive"
						size="sm"
						onClick={() => setDeleteDialog(true)}
					>
						<Trash2 className="mr-2 size-4" />
						Excluir Instancia
					</Button>
				</CardContent>
			</Card>

			<ConfirmDialog
				open={deleteDialog}
				onOpenChange={setDeleteDialog}
				title="Excluir instancia"
				description="Tem certeza que deseja excluir esta instancia? Esta acao nao pode ser desfeita. Todos os dados e historico de mensagens serao perdidos."
				confirmLabel="Excluir"
				destructive
				loading={deleteLoading}
				onConfirm={handleDelete}
			/>
		</div>
	);
}
