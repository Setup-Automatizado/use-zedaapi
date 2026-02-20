import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import {
	INSTANCE_STATUS_CONFIG,
	SUBSCRIPTION_STATUS_CONFIG,
	INVOICE_STATUS_CONFIG,
} from "@/lib/design-tokens";

interface StatusBadgeProps {
	status: string;
	type?: "instance" | "subscription" | "invoice";
	showDot?: boolean;
	className?: string;
}

export function StatusBadge({
	status,
	type = "instance",
	showDot = true,
	className,
}: StatusBadgeProps) {
	const configs = {
		instance: INSTANCE_STATUS_CONFIG,
		subscription: SUBSCRIPTION_STATUS_CONFIG,
		invoice: INVOICE_STATUS_CONFIG,
	};

	const config = configs[type][status];
	if (!config) {
		return (
			<Badge variant="secondary" className={className}>
				{status}
			</Badge>
		);
	}

	const { label, className: badgeClassName } = config;
	const dot =
		type === "instance" && "dot" in config
			? (config as { dot: string }).dot
			: undefined;

	return (
		<Badge
			variant="secondary"
			className={cn("shrink-0 font-medium", badgeClassName, className)}
		>
			{showDot && dot && (
				<span className={cn("mr-1.5 size-1.5 rounded-full", dot)} />
			)}
			{label}
		</Badge>
	);
}
