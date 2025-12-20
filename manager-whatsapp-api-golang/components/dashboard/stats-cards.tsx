import {
	CheckCircle,
	Clock,
	type LucideIcon,
	Smartphone,
	XCircle,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

export interface StatsCardsProps {
	total: number;
	connected: number;
	disconnected: number;
	pending: number;
}

interface StatCardData {
	title: string;
	value: number;
	icon: LucideIcon;
	iconColor: string;
	bgColor: string;
}

export function StatsCards({
	total,
	connected,
	disconnected,
	pending,
}: StatsCardsProps) {
	const stats: StatCardData[] = [
		{
			title: "Total Instances",
			value: total,
			icon: Smartphone,
			iconColor: "text-blue-600 dark:text-blue-400",
			bgColor: "bg-blue-500/10",
		},
		{
			title: "Connected",
			value: connected,
			icon: CheckCircle,
			iconColor: "text-emerald-600 dark:text-emerald-400",
			bgColor: "bg-emerald-500/10",
		},
		{
			title: "Disconnected",
			value: disconnected,
			icon: XCircle,
			iconColor: "text-red-600 dark:text-red-400",
			bgColor: "bg-red-500/10",
		},
		{
			title: "Pending",
			value: pending,
			icon: Clock,
			iconColor: "text-amber-600 dark:text-amber-400",
			bgColor: "bg-amber-500/10",
		},
	];

	return (
		<div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
			{stats.map((stat) => (
				<Card key={stat.title} className="relative overflow-hidden">
					<CardContent className="p-5">
						<div className="flex items-center justify-between">
							<div className="space-y-1">
								<p className="text-sm text-muted-foreground">{stat.title}</p>
								<p className="text-3xl font-bold tracking-tight tabular-nums">
									{stat.value}
								</p>
							</div>
							<div
								className={cn(
									"flex h-12 w-12 items-center justify-center rounded-full",
									stat.bgColor,
								)}
							>
								<stat.icon className={cn("h-6 w-6", stat.iconColor)} />
							</div>
						</div>
					</CardContent>
				</Card>
			))}
		</div>
	);
}
