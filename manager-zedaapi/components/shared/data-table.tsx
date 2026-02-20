"use client";

import { useState } from "react";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronLeft, ChevronRight, Search } from "lucide-react";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { useDebounce } from "@/hooks/use-debounce";
import type { LucideIcon } from "lucide-react";

export interface Column<T> {
	key: string;
	header: string;
	cell: (row: T) => React.ReactNode;
	sortable?: boolean;
	className?: string;
}

interface DataTableProps<T> {
	columns: Column<T>[];
	data: T[];
	loading?: boolean;
	searchPlaceholder?: string;
	searchKey?: keyof T;
	onSearch?: (query: string) => void;
	emptyIcon?: LucideIcon;
	emptyTitle?: string;
	emptyDescription?: string;
	emptyActionLabel?: string;
	onEmptyAction?: () => void;
	page?: number;
	pageSize?: number;
	totalCount?: number;
	onPageChange?: (page: number) => void;
	headerAction?: React.ReactNode;
}

export function DataTable<T extends { id?: string | number }>({
	columns,
	data,
	loading = false,
	searchPlaceholder = "Buscar...",
	onSearch,
	emptyIcon,
	emptyTitle = "Nenhum resultado encontrado",
	emptyDescription = "Tente ajustar os filtros de busca.",
	emptyActionLabel,
	onEmptyAction,
	page = 1,
	pageSize = 20,
	totalCount,
	onPageChange,
	headerAction,
}: DataTableProps<T>) {
	const [searchQuery, setSearchQuery] = useState("");
	const debouncedSearch = useDebounce(searchQuery, 300);

	const handleSearchChange = (value: string) => {
		setSearchQuery(value);
		onSearch?.(value);
	};

	if (loading) {
		return <TableSkeleton rows={pageSize > 10 ? 10 : pageSize} />;
	}

	const total = totalCount ?? data.length;
	const totalPages = Math.ceil(total / pageSize);
	const startItem = (page - 1) * pageSize + 1;
	const endItem = Math.min(page * pageSize, total);

	return (
		<div className="space-y-4">
			{(onSearch !== undefined || headerAction) && (
				<div className="flex items-center justify-between gap-3">
					{onSearch !== undefined && (
						<div className="relative w-full sm:w-80">
							<Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
							<Input
								placeholder={searchPlaceholder}
								value={searchQuery}
								onChange={(e) =>
									handleSearchChange(e.target.value)
								}
								className="h-9 pl-9"
							/>
						</div>
					)}
					{headerAction && (
						<div className="ml-auto shrink-0">{headerAction}</div>
					)}
				</div>
			)}

			{data.length === 0 ? (
				emptyIcon ? (
					<EmptyState
						icon={emptyIcon}
						title={emptyTitle}
						description={emptyDescription}
						actionLabel={emptyActionLabel}
						onAction={onEmptyAction}
					/>
				) : (
					<div className="flex min-h-[300px] items-center justify-center rounded-xl border border-dashed border-border">
						<div className="text-center">
							<p className="text-sm font-medium">
								{emptyTitle}
							</p>
							<p className="mt-1 text-xs text-muted-foreground">
								{emptyDescription}
							</p>
						</div>
					</div>
				)
			) : (
				<div className="overflow-hidden rounded-xl border">
					<Table>
						<TableHeader>
							<TableRow className="bg-muted/30 hover:bg-muted/30">
								{columns.map((col) => (
									<TableHead
										key={col.key}
										className={`text-xs font-medium uppercase tracking-wide ${col.className ?? ""}`}
									>
										{col.header}
									</TableHead>
								))}
							</TableRow>
						</TableHeader>
						<TableBody>
							{data.map((row, i) => (
								<TableRow
									key={row.id ?? i}
									className="transition-colors duration-100"
								>
									{columns.map((col) => (
										<TableCell
											key={col.key}
											className={col.className}
										>
											{col.cell(row)}
										</TableCell>
									))}
								</TableRow>
							))}
						</TableBody>
					</Table>
				</div>
			)}

			{total > 0 && onPageChange && (
				<div className="flex items-center justify-between text-sm text-muted-foreground">
					<span>
						{startItem}-{endItem} de {total}
					</span>
					<div className="flex gap-1">
						<Button
							variant="outline"
							size="icon-sm"
							onClick={() => onPageChange(page - 1)}
							disabled={page <= 1}
						>
							<ChevronLeft className="size-4" />
							<span className="sr-only">Pagina anterior</span>
						</Button>
						<Button
							variant="outline"
							size="icon-sm"
							onClick={() => onPageChange(page + 1)}
							disabled={page >= totalPages}
						>
							<ChevronRight className="size-4" />
							<span className="sr-only">Proxima pagina</span>
						</Button>
					</div>
				</div>
			)}
		</div>
	);
}
