import * as React from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Activity, CheckCircle, AlertTriangle, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";

export interface HealthSummaryProps {
	isHealthy: boolean;
	isDegraded: boolean;
	lastChecked: Date;
}

interface HealthStatusConfig {
	label: string;
	icon: React.ComponentType<{ className?: string }>;
	dotColor: string;
	textColor: string;
}

function getHealthConfig(
	isHealthy: boolean,
	isDegraded: boolean,
): HealthStatusConfig {
	if (isHealthy && !isDegraded) {
		return {
			label: "Online",
			icon: CheckCircle,
			dotColor: "bg-green-500",
			textColor: "text-green-600 dark:text-green-400",
		};
	}

	if (isDegraded) {
		return {
			label: "Degraded",
			icon: AlertTriangle,
			dotColor: "bg-yellow-500",
			textColor: "text-yellow-600 dark:text-yellow-400",
		};
	}

	return {
		label: "Offline",
		icon: XCircle,
		dotColor: "bg-red-500",
		textColor: "text-red-600 dark:text-red-400",
	};
}

export function HealthSummary({ isHealthy, isDegraded }: HealthSummaryProps) {
	const config = getHealthConfig(isHealthy, isDegraded);

	return (
		<Button variant="outline" size="sm" asChild>
			<Link href="/health" className="gap-2">
				<Activity className="h-4 w-4" />
				<span>API</span>
				<Badge
					variant="outline"
					className={cn(
						"ml-1 gap-1 px-1.5 py-0 text-xs font-normal",
						config.textColor,
					)}
				>
					<span
						className={cn(
							"h-1.5 w-1.5 rounded-full",
							config.dotColor,
						)}
					/>
					{config.label}
				</Badge>
			</Link>
		</Button>
	);
}
