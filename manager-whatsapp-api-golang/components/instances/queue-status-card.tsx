/**
 * Queue Status Card Component
 *
 * Displays message queue statistics with visual breakdown by status.
 * Shows total messages, success rate, and status distribution.
 *
 * @example
 * ```tsx
 * <QueueStatusCard stats={queueStats} isLoading={isLoading} />
 * ```
 */

"use client";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { QueueStats } from "@/types/queue";

export interface QueueStatusCardProps {
	/** Queue statistics data */
	stats: QueueStats | undefined;

	/** Loading state */
	isLoading: boolean;
}

export function QueueStatusCard({ stats, isLoading }: QueueStatusCardProps) {
	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle>Fila de Mensagens</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<Skeleton className="h-16 w-full" />
					<Skeleton className="h-16 w-full" />
					<Skeleton className="h-16 w-full" />
				</CardContent>
			</Card>
		);
	}

	if (!stats) {
		return (
			<Card>
				<CardHeader>
					<CardTitle>Fila de Mensagens</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="text-sm text-muted-foreground">
						Nenhum dado disponível
					</p>
				</CardContent>
			</Card>
		);
	}

	const successRate =
		stats.total > 0 ? ((stats.sent / stats.total) * 100).toFixed(1) : "0.0";

	return (
		<Card>
			<CardHeader>
				<CardTitle>Fila de Mensagens</CardTitle>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="grid grid-cols-2 gap-4">
					<div>
						<p className="text-sm font-medium text-muted-foreground">Total</p>
						<p className="text-2xl font-bold">{stats.total}</p>
					</div>
					<div>
						<p className="text-sm font-medium text-muted-foreground">
							Taxa de Sucesso
						</p>
						<p className="text-2xl font-bold">{successRate}%</p>
					</div>
				</div>

				<div className="space-y-2">
					<div className="flex justify-between items-center">
						<span className="text-sm">Pendentes</span>
						<Badge variant="outline">{stats.pending}</Badge>
					</div>
					<div className="flex justify-between items-center">
						<span className="text-sm">Processando</span>
						<Badge variant="outline">{stats.processing}</Badge>
					</div>
					<div className="flex justify-between items-center">
						<span className="text-sm">Enviadas</span>
						<Badge
							variant="outline"
							className="bg-green-50 text-green-700 border-green-200"
						>
							{stats.sent}
						</Badge>
					</div>
					<div className="flex justify-between items-center">
						<span className="text-sm">Falhas</span>
						<Badge
							variant="outline"
							className="bg-red-50 text-red-700 border-red-200"
						>
							{stats.failed}
						</Badge>
					</div>
					{stats.canceled > 0 && (
						<div className="flex justify-between items-center">
							<span className="text-sm">Canceladas</span>
							<Badge
								variant="outline"
								className="bg-gray-50 text-gray-700 border-gray-200"
							>
								{stats.canceled}
							</Badge>
						</div>
					)}
				</div>
			</CardContent>
		</Card>
	);
}
