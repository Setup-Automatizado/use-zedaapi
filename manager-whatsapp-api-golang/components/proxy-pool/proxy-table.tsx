"use client";

import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import type { PoolProxy } from "@/types/pool";
import { sanitizePoolProxyUrl } from "@/types/pool";

interface ProxyTableProps {
	proxies: PoolProxy[];
	total: number;
	isLoading?: boolean;
}

function statusBadgeVariant(
	status: string,
): "default" | "secondary" | "destructive" | "outline" {
	switch (status) {
		case "available":
			return "default";
		case "assigned":
			return "secondary";
		case "unhealthy":
			return "destructive";
		case "retired":
			return "outline";
		default:
			return "outline";
	}
}

function healthBadgeVariant(
	status: string,
): "default" | "destructive" | "outline" {
	switch (status) {
		case "healthy":
			return "default";
		case "unhealthy":
			return "destructive";
		default:
			return "outline";
	}
}

export function ProxyTable({ proxies, total, isLoading }: ProxyTableProps) {
	if (isLoading) {
		return (
			<div className="space-y-2">
				{Array.from({ length: 5 }).map((_, i) => (
					<Skeleton key={i} className="h-12 w-full" />
				))}
			</div>
		);
	}

	if (proxies.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center py-12 text-center">
				<p className="text-muted-foreground">No proxies in pool</p>
				<p className="text-sm text-muted-foreground mt-1">
					Add a provider and sync to populate the pool.
				</p>
			</div>
		);
	}

	return (
		<div>
			<div className="rounded-md border">
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>Proxy URL</TableHead>
							<TableHead>Country</TableHead>
							<TableHead>Status</TableHead>
							<TableHead>Health</TableHead>
							<TableHead className="text-right">
								Assigned
							</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{proxies.map((proxy) => (
							<TableRow key={proxy.id}>
								<TableCell className="font-mono text-xs max-w-xs truncate">
									{sanitizePoolProxyUrl(proxy.proxyUrl)}
								</TableCell>
								<TableCell>
									{proxy.countryCode ? (
										<Badge variant="outline">
											{proxy.countryCode}
										</Badge>
									) : (
										<span className="text-muted-foreground">
											-
										</span>
									)}
								</TableCell>
								<TableCell>
									<Badge
										variant={statusBadgeVariant(
											proxy.status,
										)}
									>
										{proxy.status}
									</Badge>
								</TableCell>
								<TableCell>
									<Badge
										variant={healthBadgeVariant(
											proxy.healthStatus,
										)}
									>
										{proxy.healthStatus}
									</Badge>
								</TableCell>
								<TableCell className="text-right">
									{proxy.assignedCount}/{proxy.maxAssignments}
								</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
			</div>
			<p className="text-xs text-muted-foreground mt-2">
				Showing {proxies.length} of {total} proxies
			</p>
		</div>
	);
}
