"use client";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Smartphone } from "lucide-react";

interface InstanceCardProps {
	name: string;
	status: string;
	phoneNumber?: string;
	connectedSince?: string;
}

const statusConfig: Record<
	string,
	{ label: string; dot: string; badge: string }
> = {
	connected: {
		label: "Conectado",
		dot: "bg-primary",
		badge: "bg-primary/10 text-primary",
	},
	connecting: {
		label: "Conectando",
		dot: "bg-chart-2",
		badge: "bg-chart-2/10 text-chart-2",
	},
	disconnected: {
		label: "Desconectado",
		dot: "bg-muted-foreground",
		badge: "bg-muted text-muted-foreground",
	},
	error: {
		label: "Erro",
		dot: "bg-destructive",
		badge: "bg-destructive/10 text-destructive",
	},
	banned: {
		label: "Banido",
		dot: "bg-destructive",
		badge: "bg-destructive/10 text-destructive",
	},
	creating: {
		label: "Criando",
		dot: "bg-chart-2 animate-pulse",
		badge: "bg-chart-2/10 text-chart-2",
	},
};

export function InstanceCard({
	name,
	status,
	phoneNumber,
}: InstanceCardProps) {
	const config = statusConfig[status] ?? {
		label: status,
		dot: "bg-muted-foreground",
		badge: "bg-muted text-muted-foreground",
	};

	return (
		<div className="flex items-center gap-3">
			<div className="relative">
				<div className="flex size-10 items-center justify-center rounded-lg bg-muted">
					<Smartphone className="size-4 text-muted-foreground" />
				</div>
				<div
					className={cn(
						"absolute -bottom-0.5 -right-0.5 size-2.5 rounded-full border-2 border-background",
						config.dot,
					)}
				/>
			</div>
			<div className="flex-1 min-w-0">
				<p className="truncate text-sm font-medium">{name}</p>
				{phoneNumber && (
					<p className="text-xs text-muted-foreground tabular-nums">
						{phoneNumber}
					</p>
				)}
			</div>
			<Badge
				variant="secondary"
				className={cn("shrink-0 font-medium", config.badge)}
			>
				{config.label}
			</Badge>
		</div>
	);
}
