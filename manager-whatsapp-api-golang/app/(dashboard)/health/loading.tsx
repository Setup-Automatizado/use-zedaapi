/**
 * Health Page Loading State
 *
 * Skeleton UI displayed while health data is being fetched.
 */

import * as React from "react";
import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export default function HealthLoading() {
	return (
		<div className="space-y-6">
			{/* Page Header */}
			<div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
				<PageHeader
					title="Status da API"
					description="Monitore a saude e disponibilidade do sistema"
				/>
				<Skeleton className="h-10 w-48" />
			</div>

			{/* Health Cards Grid */}
			<div className="grid gap-6 md:grid-cols-2">
				{/* Basic Health Status Card Skeleton */}
				<Card>
					<CardHeader>
						<div className="flex items-center justify-between">
							<Skeleton className="h-5 w-32" />
							<Skeleton className="h-10 w-10 rounded-full" />
						</div>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="flex items-center justify-between">
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-6 w-16" />
						</div>
						<div className="flex items-center justify-between">
							<Skeleton className="h-4 w-20" />
							<Skeleton className="h-4 w-32" />
						</div>
						<div className="flex items-center justify-between">
							<Skeleton className="h-4 w-28" />
							<Skeleton className="h-4 w-40" />
						</div>
					</CardContent>
				</Card>

				{/* Readiness Status Card Skeleton */}
				<Card>
					<CardHeader>
						<div className="flex items-center justify-between">
							<Skeleton className="h-5 w-40" />
							<Skeleton className="h-10 w-10 rounded-full" />
						</div>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="flex items-center justify-between">
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-6 w-20" />
						</div>
						<div className="flex items-center justify-between">
							<Skeleton className="h-4 w-32" />
							<Skeleton className="h-4 w-40" />
						</div>
					</CardContent>
				</Card>
			</div>

			{/* Dependency Status Skeleton */}
			<Card>
				<CardHeader>
					<Skeleton className="h-5 w-48" />
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Database Dependency */}
					<div className="flex items-center justify-between p-4 rounded-lg border">
						<div className="flex items-center gap-3">
							<Skeleton className="h-10 w-10 rounded-full" />
							<div className="space-y-2">
								<Skeleton className="h-4 w-24" />
								<Skeleton className="h-3 w-32" />
							</div>
						</div>
						<Skeleton className="h-6 w-20" />
					</div>

					{/* Redis Dependency */}
					<div className="flex items-center justify-between p-4 rounded-lg border">
						<div className="flex items-center gap-3">
							<Skeleton className="h-10 w-10 rounded-full" />
							<div className="space-y-2">
								<Skeleton className="h-4 w-24" />
								<Skeleton className="h-3 w-32" />
							</div>
						</div>
						<Skeleton className="h-6 w-20" />
					</div>

					{/* Storage Dependency */}
					<div className="flex items-center justify-between p-4 rounded-lg border">
						<div className="flex items-center gap-3">
							<Skeleton className="h-10 w-10 rounded-full" />
							<div className="space-y-2">
								<Skeleton className="h-4 w-24" />
								<Skeleton className="h-3 w-32" />
							</div>
						</div>
						<Skeleton className="h-6 w-20" />
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
