"use client";

import {
	AlertCircle,
	AlertTriangle,
	CheckCircle2,
	ChevronDown,
	Database,
	HardDrive,
	Lock,
	MessageSquare,
} from "lucide-react";
import * as React from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type {
	CircuitState,
	ComponentStatus,
	HealthStatus,
} from "@/types/health";

export interface DependencyStatusProps {
	name: string;
	status: ComponentStatus;
}

const componentIcons: Record<
	string,
	React.ComponentType<{ className?: string }>
> = {
	database: Database,
	redis: Lock,
	storage: HardDrive,
	whatsapp: MessageSquare,
};

const statusConfig: Record<
	HealthStatus,
	{
		icon: React.ComponentType<{ className?: string }>;
		color: string;
		bgColor: string;
		label: string;
		variant: "default" | "destructive" | "outline";
	}
> = {
	healthy: {
		icon: CheckCircle2,
		color: "text-green-600 dark:text-green-400",
		bgColor: "bg-green-500/10 dark:bg-green-500/20",
		label: "Healthy",
		variant: "default",
	},
	degraded: {
		icon: AlertTriangle,
		color: "text-yellow-600 dark:text-yellow-400",
		bgColor: "bg-yellow-500/10 dark:bg-yellow-500/20",
		label: "Degraded",
		variant: "outline",
	},
	unhealthy: {
		icon: AlertCircle,
		color: "text-red-600 dark:text-red-400",
		bgColor: "bg-red-500/10 dark:bg-red-500/20",
		label: "Unavailable",
		variant: "destructive",
	},
};

const circuitStateConfig: Record<
	CircuitState,
	{ label: string; color: string }
> = {
	closed: { label: "Closed", color: "text-green-600 dark:text-green-400" },
	open: { label: "Open", color: "text-red-600 dark:text-red-400" },
	"half-open": {
		label: "Half-open",
		color: "text-yellow-600 dark:text-yellow-400",
	},
};

export function DependencyStatus({ name, status }: DependencyStatusProps) {
	const [isExpanded, setIsExpanded] = React.useState(false);
	const config = statusConfig[status.status];
	const StatusIcon = config.icon;
	const ComponentIcon = componentIcons[name.toLowerCase()] || Database;
	const hasError = status.error && status.error.trim() !== "";

	return (
		<div className="border-border border-b last:border-b-0">
			<div className="flex items-center gap-3 py-3 px-1">
				{/* Component Icon */}
				<div
					className={cn(
						"size-8 rounded-lg flex items-center justify-center shrink-0",
						config.bgColor,
					)}
				>
					<ComponentIcon className={cn("size-4", config.color)} />
				</div>

				{/* Component Info */}
				<div className="flex-1 min-w-0">
					<div className="flex items-center gap-2">
						<p className="text-sm font-medium capitalize truncate">{name}</p>
						<Badge variant={config.variant} className="shrink-0">
							<StatusIcon data-icon="inline-start" className="size-3" />
							{config.label}
						</Badge>
					</div>

					<div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
						<span className="flex items-center gap-1">
							Latency: <span className="font-mono">{status.duration_ms}ms</span>
						</span>

						{status.circuit_state &&
							circuitStateConfig[status.circuit_state] && (
								<>
									<span className="text-muted-foreground/50">•</span>
									<span className="flex items-center gap-1">
										Circuit:{" "}
										<span
											className={circuitStateConfig[status.circuit_state].color}
										>
											{circuitStateConfig[status.circuit_state].label}
										</span>
									</span>
								</>
							)}
					</div>
				</div>

				{/* Expand Button (only if error) */}
				{hasError && (
					<button
						type="button"
						onClick={() => setIsExpanded(!isExpanded)}
						className="size-6 rounded-lg hover:bg-muted flex items-center justify-center shrink-0 transition-colors"
						aria-label={isExpanded ? "Hide error" : "Show error"}
					>
						<ChevronDown
							className={cn(
								"size-4 transition-transform",
								isExpanded && "rotate-180",
							)}
						/>
					</button>
				)}
			</div>

			{/* Error Details */}
			{hasError && isExpanded && (
				<div className="px-1 pb-3">
					<div className="bg-destructive/5 dark:bg-destructive/10 border border-destructive/20 rounded-lg p-3">
						<p className="text-xs font-medium text-destructive mb-1">
							Error Details
						</p>
						<p className="text-xs text-muted-foreground font-mono break-all">
							{status.error}
						</p>
					</div>
				</div>
			)}

			{/* Metadata (if available) */}
			{status.metadata &&
				Object.keys(status.metadata).length > 0 &&
				isExpanded && (
					<div className="px-1 pb-3">
						<div className="bg-muted/50 rounded-lg p-3">
							<p className="text-xs font-medium mb-2">Metadata</p>
							<dl className="space-y-1 text-xs">
								{Object.entries(status.metadata).map(([key, value]) => (
									<div key={key} className="flex gap-2">
										<dt className="text-muted-foreground font-medium capitalize">
											{key}:
										</dt>
										<dd className="font-mono text-foreground">
											{JSON.stringify(value)}
										</dd>
									</div>
								))}
							</dl>
						</div>
					</div>
				)}
		</div>
	);
}
