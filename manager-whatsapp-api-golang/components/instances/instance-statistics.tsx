/**
 * Instance Statistics Component
 *
 * Comprehensive statistics dashboard showing:
 * - Queue metrics (pending, processing, sent, failed)
 * - Status cache metrics (entries, webhooks, suppressed)
 * - General instance metrics (messages sent/received, latency, errors)
 *
 * @example
 * ```tsx
 * <InstanceStatistics instance={instance} />
 * ```
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useQueueStats } from "@/hooks/use-queue-stats";
import { useStatusCacheStats } from "@/hooks/use-status-cache-stats";
import type { Instance } from "@/types";
import { QueueStatusCard } from "./queue-status-card";

export interface InstanceStatisticsProps {
	/** Instance to show statistics for */
	instance: Instance;
}

export function InstanceStatistics({ instance }: InstanceStatisticsProps) {
	const { stats: queueStats, isLoading: queueLoading } = useQueueStats(
		instance.id,
		instance.token,
	);

	const { stats: cacheStats, isLoading: cacheLoading } = useStatusCacheStats(
		instance.id,
		instance.token,
	);

	return (
		<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{/* Queue Statistics */}
			<QueueStatusCard stats={queueStats} isLoading={queueLoading} />

			{/* Status Cache Statistics */}
			<Card>
				<CardHeader>
					<CardTitle>Status Cache</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					{cacheLoading ? (
						<>
							<Skeleton className="h-16 w-full" />
							<Skeleton className="h-16 w-full" />
						</>
					) : cacheStats ? (
						<>
							<div className="grid grid-cols-2 gap-4">
								<div>
									<p className="text-sm font-medium text-muted-foreground">
										Total de Entradas
									</p>
									<p className="text-2xl font-bold">{cacheStats.total_entries}</p>
								</div>
								<div>
									<p className="text-sm font-medium text-muted-foreground">
										Webhooks Suprimidos
									</p>
									<p className="text-2xl font-bold text-green-600">
										{cacheStats.suppressed_count}
									</p>
								</div>
							</div>
							<div className="grid grid-cols-2 gap-4">
								<div>
									<p className="text-sm font-medium text-muted-foreground">
										Pendentes
									</p>
									<p className="text-xl font-semibold">
										{cacheStats.pending_webhooks}
									</p>
								</div>
								<div>
									<p className="text-sm font-medium text-muted-foreground">
										Flushes
									</p>
									<p className="text-xl font-semibold">{cacheStats.flush_count}</p>
								</div>
							</div>
						</>
					) : (
						<p className="text-sm text-muted-foreground">
							Nenhum dado disponível
						</p>
					)}
				</CardContent>
			</Card>

			{/* Instance General Info */}
			<Card>
				<CardHeader>
					<CardTitle>Informações da Instância</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<div>
						<p className="text-sm font-medium text-muted-foreground">
							Connection Status
						</p>
						<p className="text-lg font-semibold capitalize">
							{instance.connectionStatus || "Unknown"}
						</p>
					</div>
					{instance.storeJid && (
						<div>
							<p className="text-sm font-medium text-muted-foreground">
								Store JID
							</p>
							<p className="text-lg font-mono text-xs">{instance.storeJid}</p>
						</div>
					)}
					<div>
						<p className="text-sm font-medium text-muted-foreground">
							WhatsApp Connection
						</p>
						<p className="text-lg font-semibold">
							{instance.whatsappConnected ? "Connected" : "Disconnected"}
						</p>
					</div>
					<div>
						<p className="text-sm font-medium text-muted-foreground">
							Phone Connection
						</p>
						<p className="text-lg font-semibold">
							{instance.phoneConnected ? "Connected" : "Disconnected"}
						</p>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
