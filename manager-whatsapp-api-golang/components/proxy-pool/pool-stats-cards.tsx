"use client";

import { Activity, CheckCircle, Globe, Server, XCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { PoolStats } from "@/types/pool";

interface PoolStatsCardsProps {
	stats: PoolStats | null;
	isLoading?: boolean;
}

export function PoolStatsCards({ stats, isLoading }: PoolStatsCardsProps) {
	if (isLoading || !stats) {
		return (
			<div className="grid gap-4 grid-cols-2 lg:grid-cols-5">
				{Array.from({ length: 5 }).map((_, i) => (
					<Card key={i}>
						<CardHeader className="flex flex-row items-center justify-between pb-2">
							<Skeleton className="h-4 w-20" />
							<Skeleton className="h-4 w-4" />
						</CardHeader>
						<CardContent>
							<Skeleton className="h-8 w-16" />
						</CardContent>
					</Card>
				))}
			</div>
		);
	}

	const items = [
		{
			label: "Total",
			value: stats.totalProxies,
			icon: Globe,
			color: "text-foreground",
		},
		{
			label: "Available",
			value: stats.availableProxies,
			icon: CheckCircle,
			color: "text-green-500",
		},
		{
			label: "Assigned",
			value: stats.assignedProxies,
			icon: Server,
			color: "text-blue-500",
		},
		{
			label: "Unhealthy",
			value: stats.unhealthyProxies,
			icon: XCircle,
			color: "text-red-500",
		},
		{
			label: "Assignments",
			value: stats.totalAssignments,
			icon: Activity,
			color: "text-purple-500",
		},
	];

	return (
		<div className="grid gap-4 grid-cols-2 lg:grid-cols-5">
			{items.map((item) => (
				<Card key={item.label}>
					<CardHeader className="flex flex-row items-center justify-between pb-2">
						<CardTitle className="text-sm font-medium text-muted-foreground">
							{item.label}
						</CardTitle>
						<item.icon className={`h-4 w-4 ${item.color}`} />
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">{item.value}</div>
					</CardContent>
				</Card>
			))}
		</div>
	);
}
