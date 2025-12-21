/**
 * Metric Chart Component
 *
 * Wrapper around Recharts for consistent styling and theme support.
 *
 * @module components/metrics/metric-chart
 */

"use client";

import * as React from "react";
import {
	Area,
	AreaChart,
	Bar,
	BarChart,
	CartesianGrid,
	Cell,
	Legend,
	Line,
	LineChart,
	Pie,
	PieChart,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { TAILWIND_CHART_COLORS } from "@/lib/metrics/constants";

export interface ChartDataPoint {
	[key: string]: string | number;
}

export interface ChartKey {
	key: string;
	color: string;
	label: string;
}

export interface MetricChartProps {
	/** Chart type */
	type: "line" | "bar" | "area" | "pie";
	/** Chart data */
	data: ChartDataPoint[];
	/** X-axis data key */
	xKey?: string;
	/** Y-axis data keys with colors and labels */
	yKeys: ChartKey[];
	/** Chart title */
	title?: string;
	/** Chart height in pixels */
	height?: number;
	/** Show legend */
	showLegend?: boolean;
	/** Show grid lines */
	showGrid?: boolean;
	/** Show tooltip */
	showTooltip?: boolean;
	/** Stack bars/areas */
	stacked?: boolean;
	/** Loading state */
	isLoading?: boolean;
	/** Additional CSS classes */
	className?: string;
	/** Custom colors for pie chart segments (one per data point) */
	colors?: string[];
}

export function MetricChart({
	type,
	data,
	xKey = "name",
	yKeys,
	title,
	height = 300,
	showLegend = true,
	showGrid = true,
	showTooltip = true,
	stacked = false,
	isLoading = false,
	className,
	colors,
}: MetricChartProps) {
	if (isLoading) {
		return (
			<Card className={className}>
				{title && (
					<CardHeader className="pb-2">
						<Skeleton className="h-5 w-32" />
					</CardHeader>
				)}
				<CardContent>
					<Skeleton className="w-full" style={{ height }} />
				</CardContent>
			</Card>
		);
	}

	const renderChart = () => {
		switch (type) {
			case "line":
				return (
					<LineChart data={data}>
						{showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
						<XAxis
							dataKey={xKey}
							tick={{ fontSize: 12 }}
							tickLine={false}
							axisLine={false}
							className="text-muted-foreground"
						/>
						<YAxis
							tick={{ fontSize: 12 }}
							tickLine={false}
							axisLine={false}
							className="text-muted-foreground"
						/>
						{showTooltip && <Tooltip content={<CustomTooltip />} />}
						{showLegend && <Legend />}
						{yKeys.map((yKey) => (
							<Line
								key={yKey.key}
								type="monotone"
								dataKey={yKey.key}
								name={yKey.label}
								stroke={yKey.color}
								strokeWidth={2}
								dot={false}
								activeDot={{ r: 4 }}
							/>
						))}
					</LineChart>
				);

			case "area":
				return (
					<AreaChart data={data}>
						{showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
						<XAxis
							dataKey={xKey}
							tick={{ fontSize: 12 }}
							tickLine={false}
							axisLine={false}
							className="text-muted-foreground"
						/>
						<YAxis
							tick={{ fontSize: 12 }}
							tickLine={false}
							axisLine={false}
							className="text-muted-foreground"
						/>
						{showTooltip && <Tooltip content={<CustomTooltip />} />}
						{showLegend && <Legend />}
						{yKeys.map((yKey) => (
							<Area
								key={yKey.key}
								type="monotone"
								dataKey={yKey.key}
								name={yKey.label}
								stroke={yKey.color}
								fill={yKey.color}
								fillOpacity={0.2}
								stackId={stacked ? "stack" : undefined}
							/>
						))}
					</AreaChart>
				);

			case "bar":
				return (
					<BarChart data={data}>
						{showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
						<XAxis
							dataKey={xKey}
							tick={{ fontSize: 12 }}
							tickLine={false}
							axisLine={false}
							className="text-muted-foreground"
						/>
						<YAxis
							tick={{ fontSize: 12 }}
							tickLine={false}
							axisLine={false}
							className="text-muted-foreground"
						/>
						{showTooltip && <Tooltip content={<CustomTooltip />} />}
						{showLegend && <Legend />}
						{yKeys.map((yKey) => (
							<Bar
								key={yKey.key}
								dataKey={yKey.key}
								name={yKey.label}
								fill={yKey.color}
								radius={[4, 4, 0, 0]}
								stackId={stacked ? "stack" : undefined}
							/>
						))}
					</BarChart>
				);

			case "pie":
				return (
					<PieChart>
						<Pie
							data={data}
							dataKey={yKeys[0]?.key || "value"}
							nameKey={xKey}
							cx="50%"
							cy="50%"
							outerRadius={height / 3}
							label={({ name, percent }: { name?: string; percent?: number }) =>
								`${name ?? ""}: ${((percent ?? 0) * 100).toFixed(0)}%`
							}
							labelLine={false}
						>
							{data.map((_, index) => (
								<Cell
									key={`cell-${index}`}
									fill={
										colors?.[index] ||
										yKeys[index % yKeys.length]?.color ||
										Object.values(TAILWIND_CHART_COLORS)[
											index % Object.values(TAILWIND_CHART_COLORS).length
										]
									}
								/>
							))}
						</Pie>
						{showTooltip && <Tooltip content={<CustomTooltip />} />}
						{showLegend && <Legend />}
					</PieChart>
				);

			default:
				return null;
		}
	};

	return (
		<Card className={className}>
			{title && (
				<CardHeader className="pb-2">
					<CardTitle className="text-base font-medium">{title}</CardTitle>
				</CardHeader>
			)}
			<CardContent className="pt-0">
				<ResponsiveContainer width="100%" height={height}>
					{renderChart()}
				</ResponsiveContainer>
			</CardContent>
		</Card>
	);
}

/**
 * Custom Tooltip Component
 */
function CustomTooltip({
	active,
	payload,
	label,
}: {
	active?: boolean;
	payload?: Array<{ name: string; value: number; color: string }>;
	label?: string;
}) {
	if (!active || !payload?.length) {
		return null;
	}

	return (
		<div className="rounded-lg border bg-background p-3 shadow-lg">
			{label && (
				<p className="mb-2 text-sm font-medium text-foreground">{label}</p>
			)}
			<div className="space-y-1">
				{payload.map((entry, index) => (
					<div key={index} className="flex items-center gap-2 text-sm">
						<span
							className="h-3 w-3 rounded-full"
							style={{ backgroundColor: entry.color }}
						/>
						<span className="text-muted-foreground">{entry.name}:</span>
						<span className="font-medium tabular-nums">
							{typeof entry.value === "number"
								? entry.value.toLocaleString()
								: entry.value}
						</span>
					</div>
				))}
			</div>
		</div>
	);
}

/**
 * Simple horizontal bar chart for top items
 */
export interface HorizontalBarChartProps {
	data: Array<{ name: string; value: number }>;
	title?: string;
	maxItems?: number;
	color?: string;
	isLoading?: boolean;
	className?: string;
}

export function HorizontalBarChart({
	data,
	title,
	maxItems = 5,
	color = TAILWIND_CHART_COLORS.primary,
	isLoading = false,
	className,
}: HorizontalBarChartProps) {
	const sortedData = [...data]
		.sort((a, b) => b.value - a.value)
		.slice(0, maxItems);

	const maxValue = Math.max(...sortedData.map((d) => d.value), 1);

	if (isLoading) {
		return (
			<Card className={className}>
				{title && (
					<CardHeader className="pb-2">
						<Skeleton className="h-5 w-32" />
					</CardHeader>
				)}
				<CardContent className="space-y-3">
					{Array.from({ length: maxItems }).map((_, i) => (
						<div key={i} className="space-y-1">
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-2 w-full" />
						</div>
					))}
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
			<CardContent className="space-y-3">
				{sortedData.map((item) => (
					<div key={item.name} className="space-y-1">
						<div className="flex items-center justify-between text-sm">
							<span className="truncate text-muted-foreground">{item.name}</span>
							<span className="font-medium tabular-nums">
								{item.value.toLocaleString()}
							</span>
						</div>
						<div className="h-2 w-full overflow-hidden rounded-full bg-muted">
							<div
								className="h-full rounded-full transition-all duration-300"
								style={{
									width: `${(item.value / maxValue) * 100}%`,
									backgroundColor: color,
								}}
							/>
						</div>
					</div>
				))}
				{sortedData.length === 0 && (
					<p className="text-center text-sm text-muted-foreground py-4">
						No data available
					</p>
				)}
			</CardContent>
		</Card>
	);
}
