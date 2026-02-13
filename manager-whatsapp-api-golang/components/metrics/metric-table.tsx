/**
 * Metric Table Component
 *
 * Table for displaying detailed metric breakdowns.
 *
 * @module components/metrics/metric-table
 */

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import type { HealthLevel } from "@/types/metrics";
import { StatusIndicator } from "./status-indicator";

export interface MetricTableColumn<T> {
	key: keyof T | string;
	header: string;
	align?: "left" | "center" | "right";
	format?: (value: unknown, row: T) => React.ReactNode;
	width?: string;
}

export interface MetricTableProps<T> {
	/** Table data */
	data: T[];
	/** Column definitions */
	columns: MetricTableColumn<T>[];
	/** Table title */
	title?: string;
	/** Loading state */
	isLoading?: boolean;
	/** Empty state message */
	emptyMessage?: string;
	/** Row click handler */
	onRowClick?: (row: T) => void;
	/** Additional CSS classes */
	className?: string;
}

export function MetricTable<T extends Record<string, unknown>>({
	data,
	columns,
	title,
	isLoading = false,
	emptyMessage = "No data available",
	onRowClick,
	className,
}: MetricTableProps<T>) {
	if (isLoading) {
		return (
			<Card className={className}>
				{title && (
					<CardHeader className="pb-2">
						<Skeleton className="h-5 w-32" />
					</CardHeader>
				)}
				<CardContent>
					<div className="space-y-2">
						<Skeleton className="h-10 w-full" />
						{Array.from({ length: 5 }).map((_, i) => (
							<Skeleton key={i} className="h-12 w-full" />
						))}
					</div>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card className={className}>
			{title && (
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">{title}</CardTitle>
				</CardHeader>
			)}
			<CardContent>
				<div className="overflow-x-auto">
					<Table>
						<TableHeader>
							<TableRow>
								{columns.map((column) => (
									<TableHead
										key={String(column.key)}
										className={cn(
											column.align === "center" && "text-center",
											column.align === "right" && "text-right",
										)}
										style={{ width: column.width }}
									>
										{column.header}
									</TableHead>
								))}
							</TableRow>
						</TableHeader>
						<TableBody>
							{data.length === 0 ? (
								<TableRow>
									<TableCell
										colSpan={columns.length}
										className="h-24 text-center text-muted-foreground"
									>
										{emptyMessage}
									</TableCell>
								</TableRow>
							) : (
								data.map((row, index) => (
									<TableRow
										key={index}
										className={cn(onRowClick && "cursor-pointer hover:bg-muted/50")}
										onClick={() => onRowClick?.(row)}
									>
										{columns.map((column) => {
											const value = getNestedValue(row, String(column.key));
											const formatted = column.format
												? column.format(value, row)
												: String(value ?? "-");

											return (
												<TableCell
													key={String(column.key)}
													className={cn(
														"tabular-nums",
														column.align === "center" && "text-center",
														column.align === "right" && "text-right",
													)}
												>
													{formatted}
												</TableCell>
											);
										})}
									</TableRow>
								))
							)}
						</TableBody>
					</Table>
				</div>
			</CardContent>
		</Card>
	);
}

/**
 * Get nested value from object using dot notation
 */
function getNestedValue(obj: Record<string, unknown>, path: string): unknown {
	return path.split(".").reduce((acc, part) => {
		if (acc && typeof acc === "object") {
			return (acc as Record<string, unknown>)[part];
		}
		return undefined;
	}, obj as unknown);
}

/**
 * Status cell formatter
 */
export function StatusCell({ status }: { status: HealthLevel }) {
	const labels: Record<HealthLevel, string> = {
		healthy: "Healthy",
		warning: "Warning",
		critical: "Critical",
	};

	return (
		<div className="flex items-center gap-2">
			<StatusIndicator status={status} size="sm" />
			<span className="text-sm">{labels[status]}</span>
		</div>
	);
}

/**
 * Circuit breaker state cell
 */
export function CircuitBreakerCell({ state }: { state: string }) {
	const status: HealthLevel =
		state === "closed"
			? "healthy"
			: state === "half-open"
				? "warning"
				: state === "open"
					? "critical"
					: "warning";

	const labels: Record<string, string> = {
		closed: "Closed",
		"half-open": "Half-Open",
		open: "Open",
		unknown: "Unknown",
	};

	return (
		<div className="flex items-center gap-2">
			<StatusIndicator status={status} size="sm" />
			<span className="text-sm capitalize">{labels[state] || state}</span>
		</div>
	);
}

/**
 * Number cell with formatting
 */
export function NumberCell({
	value,
	decimals = 0,
	prefix,
	suffix,
}: {
	value: number;
	decimals?: number;
	prefix?: string;
	suffix?: string;
}) {
	const formatted = Number.isFinite(value)
		? value.toLocaleString(undefined, {
				minimumFractionDigits: decimals,
				maximumFractionDigits: decimals,
			})
		: "-";

	return (
		<span className="tabular-nums">
			{prefix}
			{formatted}
			{suffix}
		</span>
	);
}

/**
 * Duration cell
 */
export function DurationCell({ ms }: { ms: number }) {
	if (!Number.isFinite(ms)) return <span>-</span>;

	if (ms < 1) {
		return <span className="tabular-nums">{(ms * 1000).toFixed(0)}us</span>;
	}
	if (ms < 1000) {
		return <span className="tabular-nums">{ms.toFixed(1)}ms</span>;
	}
	if (ms < 60000) {
		return <span className="tabular-nums">{(ms / 1000).toFixed(2)}s</span>;
	}

	return <span className="tabular-nums">{(ms / 60000).toFixed(1)}m</span>;
}

/**
 * Bytes cell
 */
export function BytesCell({ bytes }: { bytes: number }) {
	if (!Number.isFinite(bytes)) return <span>-</span>;

	const units = ["B", "KB", "MB", "GB", "TB"];
	let value = bytes;
	let unitIndex = 0;

	while (value >= 1024 && unitIndex < units.length - 1) {
		value /= 1024;
		unitIndex++;
	}

	return (
		<span className="tabular-nums">
			{value.toFixed(unitIndex > 0 ? 1 : 0)} {units[unitIndex]}
		</span>
	);
}

/**
 * Percentage cell
 */
export function PercentageCell({
	value,
	decimals = 1,
}: {
	value: number;
	decimals?: number;
}) {
	if (!Number.isFinite(value)) return <span>-</span>;

	return <span className="tabular-nums">{value.toFixed(decimals)}%</span>;
}
