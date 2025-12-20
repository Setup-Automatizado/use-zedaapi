import * as React from "react";
import { Badge } from "@/components/ui/badge";
import {
	INSTANCE_STATUS,
	type InstanceStatus,
	STATUS_COLORS,
} from "@/lib/constants";
import { cn } from "@/lib/utils";

export interface StatusBadgeProps {
	status: InstanceStatus;
	label?: string;
	showDot?: boolean;
	className?: string;
}

const statusLabels: Record<InstanceStatus, string> = {
	[INSTANCE_STATUS.CONNECTED]: "Connected",
	[INSTANCE_STATUS.DISCONNECTED]: "Disconnected",
	[INSTANCE_STATUS.PENDING]: "Pending",
	[INSTANCE_STATUS.ERROR]: "Error",
};

export function StatusBadge({
	status,
	label,
	showDot = true,
	className,
}: StatusBadgeProps) {
	const colors = STATUS_COLORS[status] || STATUS_COLORS.error;
	const displayLabel = label || statusLabels[status];

	return (
		<Badge
			variant="outline"
			className={cn(
				"gap-1.5",
				colors.bg,
				colors.border,
				colors.text,
				className,
			)}
		>
			{showDot && (
				<span
					className={cn("h-1.5 w-1.5 rounded-full", colors.dot)}
					aria-hidden="true"
				/>
			)}
			{displayLabel}
		</Badge>
	);
}
