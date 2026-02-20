"use client";

import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Smartphone } from "lucide-react";
import { cn } from "@/lib/utils";
import { EmptyState } from "@/components/shared/empty-state";

interface InstanceSummary {
	id: string;
	name: string;
	status: string;
	phoneNumber?: string;
}

interface InstanceOverviewProps {
	instances: InstanceSummary[];
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
	creating: {
		label: "Criando",
		className: "bg-chart-2/10 text-chart-2",
		dot: "bg-chart-2",
	},
};

export function InstanceOverview({ instances }: InstanceOverviewProps) {
	if (instances.length === 0) {
		return (
			<EmptyState
				icon={Smartphone}
				title="Nenhuma instancia criada"
				description="Crie sua primeira instancia WhatsApp para comecar a enviar mensagens."
				actionLabel="Criar Instancia"
				onAction={() => {}}
			/>
		);
	}

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h2 className="text-base font-semibold">Suas Instancias</h2>
				<Button asChild size="sm" variant="outline">
					<Link href="/dashboard/instances">Ver todas</Link>
				</Button>
			</div>
			<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{instances.slice(0, 6).map((instance) => {
					const config = statusConfig[instance.status] ?? {
						label: instance.status,
						className: "bg-muted text-muted-foreground",
						dot: "bg-muted-foreground",
					};

					return (
						<Link
							key={instance.id}
							href={`/dashboard/instances/${instance.id}`}
						>
							<Card className="transition-all duration-200 hover:shadow-sm hover:bg-accent/50 cursor-pointer">
								<CardContent className="flex items-center gap-3 py-3">
									<div className="relative">
										<div className="flex size-9 items-center justify-center rounded-lg bg-muted">
											<Smartphone className="size-4 text-muted-foreground" />
										</div>
										<div
											className={cn(
												"absolute -bottom-0.5 -right-0.5 size-2 rounded-full border-2 border-card",
												config.dot,
											)}
										/>
									</div>
									<div className="flex-1 min-w-0">
										<p className="truncate text-sm font-medium">
											{instance.name}
										</p>
										{instance.phoneNumber && (
											<p className="text-xs text-muted-foreground tabular-nums">
												{instance.phoneNumber}
											</p>
										)}
									</div>
									<Badge
										variant="secondary"
										className={cn(
											"shrink-0 text-[10px] font-medium",
											config.className,
										)}
									>
										{config.label}
									</Badge>
								</CardContent>
							</Card>
						</Link>
					);
				})}
			</div>
		</div>
	);
}
