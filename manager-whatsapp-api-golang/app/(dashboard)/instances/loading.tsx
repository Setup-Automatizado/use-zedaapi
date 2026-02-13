/**
 * Instances Page Loading State
 *
 * Skeleton UI displayed while instances data is being fetched.
 */

import { Plus } from "lucide-react";
import * as React from "react";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

export default function InstancesLoading() {
	return (
		<div className="space-y-6">
			{/* Page Header */}
			<PageHeader
				title="Instancias"
				description="Gerencie suas instancias do WhatsApp"
				action={
					<Button disabled>
						<Plus className="h-4 w-4" />
						Nova Instancia
					</Button>
				}
			/>

			{/* Filters Skeleton */}
			<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
				<Skeleton className="h-10 w-full max-w-md" />
				<div className="flex items-center gap-2">
					<Skeleton className="h-10 w-[180px]" />
				</div>
			</div>

			{/* Table Skeleton */}
			<div className="rounded-lg border bg-card">
				{/* Table Header */}
				<div className="border-b bg-muted/50 p-4">
					<div className="flex items-center justify-between">
						<div className="flex gap-4 flex-1">
							<Skeleton className="h-4 w-32" />
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-4 w-28" />
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-4 w-28" />
						</div>
						<Skeleton className="h-4 w-8" />
					</div>
				</div>

				{/* Table Rows */}
				{Array.from({ length: 5 }).map((_, i) => (
					<div
						key={i}
						className="flex items-center justify-between border-b p-4 last:border-0"
					>
						<div className="flex gap-4 flex-1">
							<div className="space-y-2 w-32">
								<Skeleton className="h-4 w-28" />
								<Skeleton className="h-3 w-24" />
							</div>
							<Skeleton className="h-6 w-24" />
							<Skeleton className="h-4 w-28" />
							<Skeleton className="h-6 w-20" />
							<Skeleton className="h-4 w-28" />
						</div>
						<Skeleton className="h-8 w-8 rounded-full" />
					</div>
				))}
			</div>

			{/* Pagination Skeleton */}
			<div className="flex justify-center">
				<div className="flex items-center gap-2">
					<Skeleton className="h-10 w-24" />
					<Skeleton className="h-10 w-10" />
					<Skeleton className="h-10 w-10" />
					<Skeleton className="h-10 w-10" />
					<Skeleton className="h-10 w-24" />
				</div>
			</div>
		</div>
	);
}
