import {
	CheckCircle,
	Clock,
	type LucideIcon,
	Smartphone,
	XCircle,
} from "lucide-react";
import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
	iconBg: string;
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
			iconColor: "text-primary",
			iconBg: "bg-primary/10",
		},
		{
			title: "Connected",
			value: connected,
			icon: CheckCircle,
			iconColor: "text-green-600 dark:text-green-500",
			iconBg: "bg-green-50 dark:bg-green-950/20",
		},
		{
			title: "Disconnected",
			value: disconnected,
			icon: XCircle,
			iconColor: "text-red-600 dark:text-red-500",
			iconBg: "bg-red-50 dark:bg-red-950/20",
		},
		{
			title: "Pending",
			value: pending,
			icon: Clock,
			iconColor: "text-yellow-600 dark:text-yellow-500",
			iconBg: "bg-yellow-50 dark:bg-yellow-950/20",
		},
	];

	return (
		<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{stats.map((stat) => (
				<Card key={stat.title} size="sm">
					<CardHeader>
						<div className="flex items-center justify-between">
							<CardTitle className="text-sm font-medium text-muted-foreground">
								{stat.title}
							</CardTitle>
							<div
								className={cn(
									"flex h-9 w-9 items-center justify-center rounded-2xl",
									stat.iconBg,
								)}
							>
								<stat.icon className={cn("h-5 w-5", stat.iconColor)} />
							</div>
						</div>
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{stat.value}</div>
					</CardContent>
				</Card>
			))}
		</div>
	);
}
