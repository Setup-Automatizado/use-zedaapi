import * as React from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export interface InstanceStatusBadgeProps {
	connected: boolean;
	smartphoneConnected?: boolean;
	className?: string;
}

export function InstanceStatusBadge({
	connected,
	smartphoneConnected,
	className,
}: InstanceStatusBadgeProps) {
	const isConnected = connected && smartphoneConnected !== false;

	const statusConfig = isConnected
		? {
				label: "Connected",
				bg: "bg-green-50 dark:bg-green-950/20",
				border: "border-green-200 dark:border-green-800",
				text: "text-green-700 dark:text-green-400",
				dot: "bg-green-500",
			}
		: {
				label: "Disconnected",
				bg: "bg-red-50 dark:bg-red-950/20",
				border: "border-red-200 dark:border-red-800",
				text: "text-red-700 dark:text-red-400",
				dot: "bg-red-500",
			};

	return (
		<Badge
			variant="outline"
			className={cn(
				"gap-1.5",
				statusConfig.bg,
				statusConfig.border,
				statusConfig.text,
				className,
			)}
		>
			<span
				className={cn("h-1.5 w-1.5 rounded-full", statusConfig.dot)}
				aria-hidden="true"
			/>
			{statusConfig.label}
		</Badge>
	);
}
