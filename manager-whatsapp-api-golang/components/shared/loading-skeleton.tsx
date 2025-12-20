import * as React from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

export function CardSkeleton({ className }: { className?: string }) {
	return (
		<div className={cn("space-y-3 rounded-lg border p-6", className)}>
			<Skeleton className="h-4 w-1/3" />
			<Skeleton className="h-8 w-2/3" />
			<Skeleton className="h-4 w-full" />
		</div>
	);
}

export function TableRowSkeleton({ columns = 4 }: { columns?: number }) {
	return (
		<tr className="border-b">
			{Array.from({ length: columns }).map((_, i) => (
				<td key={i} className="p-3">
					<Skeleton className="h-4 w-full" />
				</td>
			))}
		</tr>
	);
}

export function TableSkeleton({
	rows = 5,
	columns = 4,
}: {
	rows?: number;
	columns?: number;
}) {
	return (
		<div className="rounded-lg border">
			<div className="overflow-x-auto">
				<table className="w-full">
					<thead className="border-b bg-muted/50">
						<tr>
							{Array.from({ length: columns }).map((_, i) => (
								<th key={i} className="p-3 text-left">
									<Skeleton className="h-4 w-24" />
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{Array.from({ length: rows }).map((_, i) => (
							<TableRowSkeleton key={i} columns={columns} />
						))}
					</tbody>
				</table>
			</div>
		</div>
	);
}

export function InstanceCardSkeleton() {
	return (
		<div className="space-y-4 rounded-lg border p-6">
			<div className="flex items-start justify-between">
				<div className="space-y-2">
					<Skeleton className="h-5 w-32" />
					<Skeleton className="h-4 w-48" />
				</div>
				<Skeleton className="h-6 w-20 rounded-full" />
			</div>
			<div className="space-y-2">
				<Skeleton className="h-4 w-full" />
				<Skeleton className="h-4 w-3/4" />
			</div>
			<div className="flex gap-2">
				<Skeleton className="h-9 w-24" />
				<Skeleton className="h-9 w-24" />
			</div>
		</div>
	);
}

export function StatsSkeleton() {
	return (
		<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
			{Array.from({ length: 4 }).map((_, i) => (
				<div key={i} className="space-y-3 rounded-lg border p-6">
					<Skeleton className="h-4 w-24" />
					<Skeleton className="h-8 w-16" />
					<Skeleton className="h-3 w-32" />
				</div>
			))}
		</div>
	);
}

export function FormSkeleton() {
	return (
		<div className="space-y-6">
			<div className="space-y-2">
				<Skeleton className="h-4 w-24" />
				<Skeleton className="h-10 w-full" />
			</div>
			<div className="space-y-2">
				<Skeleton className="h-4 w-24" />
				<Skeleton className="h-10 w-full" />
			</div>
			<div className="space-y-2">
				<Skeleton className="h-4 w-24" />
				<Skeleton className="h-24 w-full" />
			</div>
			<div className="flex justify-end gap-2">
				<Skeleton className="h-10 w-24" />
				<Skeleton className="h-10 w-24" />
			</div>
		</div>
	);
}
