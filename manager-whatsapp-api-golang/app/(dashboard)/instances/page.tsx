/**
 * Instances List Page
 *
 * Displays paginated list of WhatsApp instances with filtering,
 * searching, and action capabilities.
 */

"use client";

import { AlertCircle, Smartphone } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import * as React from "react";
import {
	CreateInstanceButton,
	InstanceCard,
	InstanceFilters,
	InstanceTable,
} from "@/components/instances";
import { EmptyState } from "@/components/shared/empty-state";
import { PageHeader } from "@/components/shared/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
	Pagination,
	PaginationContent,
	PaginationItem,
	PaginationLink,
	PaginationNext,
	PaginationPrevious,
} from "@/components/ui/pagination";
import { useInstancesWithDevice } from "@/hooks/use-instances-with-device";
import { useMediaQuery } from "@/hooks/use-media-query";
import type { Instance } from "@/types";

const DEFAULT_PAGE_SIZE = 10;

export default function InstancesPage() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const isMobile = useMediaQuery("(max-width: 768px)");

	// Get filters from URL params
	const [filters, setFilters] = React.useState({
		page: Number(searchParams.get("page")) || 1,
		pageSize: Number(searchParams.get("pageSize")) || DEFAULT_PAGE_SIZE,
		query: searchParams.get("query") || "",
		status:
			(searchParams.get("status") as "all" | "connected" | "disconnected") ||
			"all",
	});

	// Fetch instances with filters and device info
	const { instances, deviceMap, pagination, isLoading, error } =
		useInstancesWithDevice(filters);

	// Update URL params when filters change
	const updateFilters = React.useCallback(
		(updates: Partial<typeof filters>) => {
			const newFilters = { ...filters, ...updates };

			// Reset to page 1 when changing filters (except page itself)
			if (updates.query !== undefined || updates.status !== undefined) {
				newFilters.page = 1;
			}

			setFilters(newFilters);

			// Update URL
			const params = new URLSearchParams();
			if (newFilters.page > 1) params.set("page", newFilters.page.toString());
			if (newFilters.pageSize !== DEFAULT_PAGE_SIZE)
				params.set("pageSize", newFilters.pageSize.toString());
			if (newFilters.query) params.set("query", newFilters.query);
			if (newFilters.status !== "all") params.set("status", newFilters.status);

			const queryString = params.toString();
			router.replace(`/instances${queryString ? `?${queryString}` : ""}`, {
				scroll: false,
			});
		},
		[filters, router],
	);

	// Handle instance actions
	const handleRestart = async (instance: Instance) => {
		// TODO: Implement restart action
		console.log("Restart instance:", instance.id);
	};

	const handleDisconnect = async (instance: Instance) => {
		// TODO: Implement disconnect action
		console.log("Disconnect instance:", instance.id);
	};

	const handleDelete = async (instance: Instance) => {
		// TODO: Implement delete action with confirmation
		console.log("Delete instance:", instance.id);
	};

	// Show error state
	if (error) {
		return (
			<div className="space-y-6">
				<PageHeader
					title="Instances"
					description="Manage your WhatsApp instances"
					action={<CreateInstanceButton />}
				/>
				<Alert variant="destructive">
					<AlertCircle className="h-4 w-4" />
					<AlertTitle>Error loading instances</AlertTitle>
					<AlertDescription>
						{error.message ||
							"Failed to load instances list. Please try again."}
					</AlertDescription>
				</Alert>
			</div>
		);
	}

	// Show empty state when no instances and no filters
	const hasNoInstances = !isLoading && (!instances || instances.length === 0);
	const hasNoFilters = filters.query === "" && filters.status === "all";

	if (hasNoInstances && hasNoFilters) {
		return (
			<div className="space-y-6">
				<PageHeader
					title="Instances"
					description="Manage your WhatsApp instances"
					action={<CreateInstanceButton />}
				/>
				<EmptyState
					icon={<Smartphone className="h-8 w-8 text-muted-foreground" />}
					title="No instances created"
					description="Create your first instance to start using the WhatsApp API"
					action={<CreateInstanceButton />}
				/>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<PageHeader
				title="Instances"
				description="Manage your WhatsApp instances"
				action={<CreateInstanceButton />}
			/>

			{/* Filters */}
			<InstanceFilters
				status={filters.status}
				query={filters.query}
				onStatusChange={(status) =>
					updateFilters({
						status: status as "all" | "connected" | "disconnected",
					})
				}
				onQueryChange={(query) => updateFilters({ query })}
				onClearFilters={() => updateFilters({ status: "all", query: "" })}
			/>

			{/* Table (Desktop) / Cards (Mobile) */}
			{isMobile ? (
				<div className="space-y-4">
					{isLoading ? (
						Array.from({ length: 3 }).map((_, i) => (
							<div key={i} className="h-48 animate-pulse rounded-lg bg-muted" />
						))
					) : instances && instances.length > 0 ? (
						instances.map((instance) => (
							<InstanceCard
								key={instance.id}
								instance={instance}
								deviceInfo={deviceMap[instance.instanceId]}
								onRestart={() => handleRestart(instance)}
								onDisconnect={() => handleDisconnect(instance)}
								onDelete={() => handleDelete(instance)}
							/>
						))
					) : (
						<EmptyState
							icon={<Smartphone className="h-8 w-8 text-muted-foreground" />}
							title="No instances found"
							description="Try adjusting the filters or create a new instance"
							action={<CreateInstanceButton />}
						/>
					)}
				</div>
			) : (
				<InstanceTable
					instances={instances || []}
					deviceMap={deviceMap}
					isLoading={isLoading}
					onRestart={handleRestart}
					onDisconnect={handleDisconnect}
					onDelete={handleDelete}
				/>
			)}

			{/* Pagination */}
			{pagination && pagination.totalPage > 1 && (
				<div className="flex justify-center">
					<Pagination>
						<PaginationContent>
							<PaginationItem>
								<PaginationPrevious
									size="default"
									onClick={() =>
										updateFilters({
											page: Math.max(1, filters.page - 1),
										})
									}
									className={
										filters.page === 1
											? "pointer-events-none opacity-50"
											: "cursor-pointer"
									}
								/>
							</PaginationItem>

							{Array.from(
								{ length: Math.min(5, pagination.totalPage) },
								(_, i) => {
									// Show first page, last page, current page, and pages around current
									const pageNumber = i + 1;
									const shouldShow =
										pageNumber === 1 ||
										pageNumber === pagination.totalPage ||
										Math.abs(pageNumber - filters.page) <= 1;

									if (!shouldShow) return null;

									return (
										<PaginationItem key={pageNumber}>
											<PaginationLink
												size="icon"
												onClick={() =>
													updateFilters({
														page: pageNumber,
													})
												}
												isActive={pageNumber === filters.page}
												className="cursor-pointer"
											>
												{pageNumber}
											</PaginationLink>
										</PaginationItem>
									);
								},
							)}

							<PaginationItem>
								<PaginationNext
									size="default"
									onClick={() =>
										updateFilters({
											page: Math.min(pagination.totalPage, filters.page + 1),
										})
									}
									className={
										filters.page === pagination.totalPage
											? "pointer-events-none opacity-50"
											: "cursor-pointer"
									}
								/>
							</PaginationItem>
						</PaginationContent>
					</Pagination>
				</div>
			)}

			{/* Results count */}
			{pagination && (
				<p className="text-center text-sm text-muted-foreground">
					Showing {instances?.length || 0} of {pagination.total} instance(s)
				</p>
			)}
		</div>
	);
}
