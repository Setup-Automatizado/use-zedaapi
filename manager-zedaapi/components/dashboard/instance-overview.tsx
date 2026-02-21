"use client";

import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Smartphone } from "lucide-react";
import { cn } from "@/lib/utils";
import { EmptyState } from "@/components/shared/empty-state";
import { StatusBadge } from "@/components/shared/status-badge";
import { INSTANCE_STATUS_CONFIG } from "@/lib/design-tokens";

interface InstanceSummary {
	id: string;
	name: string;
	status: string;
	phoneNumber?: string;
}

interface InstanceOverviewProps {
	instances: InstanceSummary[];
}

export function InstanceOverview({ instances }: InstanceOverviewProps) {
	if (instances.length === 0) {
		return (
			<EmptyState
				icon={Smartphone}
				title="Nenhuma instância criada ainda"
				description="Crie sua primeira instância WhatsApp para começar a enviar mensagens."
				actionLabel="Criar Instância"
				onAction={() => {}}
			/>
		);
	}

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h2 className="text-base font-semibold">Suas Instâncias</h2>
				<Button asChild size="sm" variant="outline">
					<Link href="/instancias">Ver todas</Link>
				</Button>
			</div>
			<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
				{instances.slice(0, 6).map((instance) => (
					<Link key={instance.id} href={`/instancias/${instance.id}`}>
						<Card className="transition-all duration-200 hover:shadow-sm hover:bg-accent/50 cursor-pointer">
							<CardContent className="flex items-center gap-3 py-3">
								<div className="relative">
									<div className="flex size-9 items-center justify-center rounded-lg bg-muted">
										<Smartphone className="size-4 text-muted-foreground" />
									</div>
									<div
										className={cn(
											"absolute -bottom-0.5 -right-0.5 size-2 rounded-full border-2 border-card",
											INSTANCE_STATUS_CONFIG[
												instance.status
											]?.dot ?? "bg-muted-foreground",
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
								<StatusBadge
									status={instance.status}
									type="instance"
									showDot={false}
									className="text-[10px]"
								/>
							</CardContent>
						</Card>
					</Link>
				))}
			</div>
		</div>
	);
}
