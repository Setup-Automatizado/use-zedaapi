"use client";

import * as React from "react";
import { BarChart3, TrendingUp, Zap } from "lucide-react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

/**
 * Placeholder for future Prometheus metrics integration
 *
 * This component is designed to display real-time metrics from Prometheus/Grafana:
 * - Request latency (p50, p95, p99)
 * - Request throughput (req/s)
 * - Error rates
 * - Active connections
 * - Queue backlog
 *
 * Future implementation will integrate with:
 * - Prometheus HTTP API for metric queries
 * - Chart libraries (recharts, visx, or nivo)
 * - Time-series data visualization
 */

export interface MetricsDisplayProps {
	className?: string;
}

export function MetricsDisplay({ className }: MetricsDisplayProps) {
	return (
		<Card className={className}>
			<CardHeader>
				<CardTitle>Metricas de Performance</CardTitle>
				<CardDescription>
					Metricas em tempo real do Prometheus (em breve)
				</CardDescription>
			</CardHeader>
			<CardContent>
				<PlaceholderMetrics />
			</CardContent>
		</Card>
	);
}

function PlaceholderMetrics() {
	return (
		<div className="space-y-4">
			{/* Metric Categories */}
			<div className="grid gap-3 sm:grid-cols-3">
				<MetricPlaceholder
					icon={Zap}
					label="Latencia Media"
					description="p95 latency"
					color="blue"
				/>
				<MetricPlaceholder
					icon={TrendingUp}
					label="Throughput"
					description="requests/sec"
					color="green"
				/>
				<MetricPlaceholder
					icon={BarChart3}
					label="Taxa de Erro"
					description="error rate %"
					color="red"
				/>
			</div>

			{/* Info Box */}
			<div className="rounded-lg border border-dashed border-muted-foreground/30 p-4 text-center">
				<BarChart3 className="size-8 mx-auto mb-2 text-muted-foreground/50" />
				<p className="text-sm font-medium text-muted-foreground mb-1">
					Metricas do Prometheus
				</p>
				<p className="text-xs text-muted-foreground">
					Integracao com Prometheus e visualizacao de graficos em
					desenvolvimento. Esta area exibira metricas de latencia,
					throughput, taxa de erro e outros indicadores de
					performance.
				</p>
			</div>

			{/* Future Features List */}
			<div className="space-y-2">
				<p className="text-xs font-medium text-muted-foreground">
					Proximas funcionalidades:
				</p>
				<ul className="space-y-1 text-xs text-muted-foreground">
					<li className="flex items-start gap-2">
						<span className="text-primary mt-0.5">•</span>
						<span>
							Graficos de latencia (p50, p95, p99) em tempo real
						</span>
					</li>
					<li className="flex items-start gap-2">
						<span className="text-primary mt-0.5">•</span>
						<span>
							Metricas de fila de mensagens (backlog, throughput)
						</span>
					</li>
					<li className="flex items-start gap-2">
						<span className="text-primary mt-0.5">•</span>
						<span>Historico de conexoes e eventos do WhatsApp</span>
					</li>
					<li className="flex items-start gap-2">
						<span className="text-primary mt-0.5">•</span>
						<span>Alertas e notificacoes de threshold</span>
					</li>
					<li className="flex items-start gap-2">
						<span className="text-primary mt-0.5">•</span>
						<span>Exportacao de dados para CSV/JSON</span>
					</li>
				</ul>
			</div>
		</div>
	);
}

function MetricPlaceholder({
	icon: Icon,
	label,
	description,
	color,
}: {
	icon: React.ComponentType<{ className?: string }>;
	label: string;
	description: string;
	color: "blue" | "green" | "red";
}) {
	const colorClasses = {
		blue: "bg-blue-50/50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 text-blue-600 dark:text-blue-400",
		green: "bg-green-50/50 dark:bg-green-950/20 border-green-200 dark:border-green-800 text-green-600 dark:text-green-400",
		red: "bg-red-50/50 dark:bg-red-950/20 border-red-200 dark:border-red-800 text-red-600 dark:text-red-400",
	};

	return (
		<div className={cn("rounded-lg border p-3", colorClasses[color])}>
			<div className="flex items-center gap-2 mb-2">
				<Icon className="size-4" />
				<p className="text-xs font-medium">{label}</p>
			</div>
			<div className="flex items-baseline gap-1">
				<span className="text-2xl font-bold">--</span>
				<span className="text-xs opacity-70">{description}</span>
			</div>
		</div>
	);
}

/**
 * Future: Individual metric card for dashboard grid
 */
export interface MetricCardProps {
	label: string;
	value: number | string;
	unit?: string;
	trend?: "up" | "down" | "neutral";
	icon?: React.ComponentType<{ className?: string }>;
	className?: string;
}

export function MetricCard({
	label,
	value,
	unit,
	trend,
	icon: Icon,
	className,
}: MetricCardProps) {
	return (
		<Card className={className}>
			<CardContent className="p-4">
				<div className="flex items-center justify-between mb-2">
					<p className="text-xs text-muted-foreground">{label}</p>
					{Icon && <Icon className="size-4 text-muted-foreground" />}
				</div>
				<div className="flex items-baseline gap-1">
					<span className="text-2xl font-bold">{value}</span>
					{unit && (
						<span className="text-sm text-muted-foreground">
							{unit}
						</span>
					)}
				</div>
				{trend && (
					<div
						className={cn(
							"text-xs mt-1",
							trend === "up" &&
								"text-green-600 dark:text-green-400",
							trend === "down" &&
								"text-red-600 dark:text-red-400",
							trend === "neutral" && "text-muted-foreground",
						)}
					>
						{trend === "up" && "↑"}
						{trend === "down" && "↓"}
						{trend === "neutral" && "→"}
					</div>
				)}
			</CardContent>
		</Card>
	);
}
