"use client";

import * as React from "react";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { TableRowSkeleton } from "./loading-skeleton";
import { EmptyState } from "./empty-state";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export interface Column<T> {
	key: string;
	label: string;
	render?: (item: T) => React.ReactNode;
	className?: string;
}

export interface DataTableProps<T> {
	columns: Column<T>[];
	data: T[];
	isLoading?: boolean;
	emptyMessage?: string;
	emptyDescription?: string;
	emptyIcon?: React.ReactNode;
	getRowKey?: (item: T, index: number) => string;
	pagination?: {
		page: number;
		pageSize: number;
		total: number;
		onPageChange: (page: number) => void;
	};
	className?: string;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function DataTable<T extends Record<string, any>>({
	columns,
	data,
	isLoading = false,
	emptyMessage = "No data found",
	emptyDescription,
	emptyIcon,
	getRowKey = (item, index) => item.id?.toString() || index.toString(),
	pagination,
	className,
}: DataTableProps<T>) {
	const isEmpty = !isLoading && data.length === 0;
	const showPagination = pagination && pagination.total > pagination.pageSize;

	const totalPages = pagination
		? Math.ceil(pagination.total / pagination.pageSize)
		: 0;

	const canGoPrevious = pagination && pagination.page > 1;
	const canGoNext = pagination && pagination.page < totalPages;

	const startItem = pagination
		? (pagination.page - 1) * pagination.pageSize + 1
		: 0;
	const endItem = pagination
		? Math.min(pagination.page * pagination.pageSize, pagination.total)
		: 0;

	if (isEmpty) {
		return (
			<EmptyState
				title={emptyMessage}
				description={emptyDescription}
				icon={emptyIcon}
				className={className}
			/>
		);
	}

	return (
		<div className={cn("space-y-4", className)}>
			<div className="rounded-lg border">
				<Table>
					<TableHeader>
						<TableRow>
							{columns.map((column) => (
								<TableHead
									key={column.key}
									className={column.className}
								>
									{column.label}
								</TableHead>
							))}
						</TableRow>
					</TableHeader>
					<TableBody>
						{isLoading
							? Array.from({ length: 5 }).map((_, index) => (
									<TableRowSkeleton
										key={index}
										columns={columns.length}
									/>
								))
							: data.map((item, index) => (
									<TableRow key={getRowKey(item, index)}>
										{columns.map((column) => (
											<TableCell
												key={column.key}
												className={column.className}
											>
												{column.render
													? column.render(item)
													: item[
															column.key
														]?.toString() || "-"}
											</TableCell>
										))}
									</TableRow>
								))}
					</TableBody>
				</Table>
			</div>

			{showPagination && !isLoading && (
				<div className="flex items-center justify-between">
					<div className="text-sm text-muted-foreground">
						Showing {startItem} to {endItem} of {pagination.total}{" "}
						results
					</div>
					<div className="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							onClick={() =>
								pagination.onPageChange(pagination.page - 1)
							}
							disabled={!canGoPrevious}
						>
							<ChevronLeft className="h-4 w-4" />
							Previous
						</Button>
						<div className="flex items-center gap-1">
							<span className="text-sm">
								Page {pagination.page} of {totalPages}
							</span>
						</div>
						<Button
							variant="outline"
							size="sm"
							onClick={() =>
								pagination.onPageChange(pagination.page + 1)
							}
							disabled={!canGoNext}
						>
							Next
							<ChevronRight className="h-4 w-4" />
						</Button>
					</div>
				</div>
			)}
		</div>
	);
}
