"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
	HealthStatusCard,
	ReadinessCard,
	DependencyStatus,
	AutoRefreshIndicator,
	AutoRefreshIndicatorCompact,
	MetricsDisplay,
} from "@/components/health";
import type {
	HealthResponse,
	ReadinessResponse,
} from "@/types/health";

/**
 * Health Components Playground
 *
 * Componente para testar visualmente todos os estados dos componentes de health.
 * Util durante desenvolvimento e para demonstracoes.
 *
 * Uso:
 * 1. Crie uma rota em app/playground/health/page.tsx
 * 2. Importe: import { HealthPlayground } from "@/components/health/health-playground"
 * 3. Renderize: export default function Page() { return <HealthPlayground /> }
 */

export function HealthPlayground() {
	const [scenario, setScenario] = React.useState<
		"healthy" | "degraded" | "unhealthy" | "loading" | "error"
	>("healthy");

	const [lastRefresh, setLastRefresh] = React.useState(new Date());
	const [isRefreshing, setIsRefreshing] = React.useState(false);

	const handleRefresh = () => {
		setIsRefreshing(true);
		setTimeout(() => {
			setLastRefresh(new Date());
			setIsRefreshing(false);
		}, 1000);
	};

	const mockData = getMockData(scenario);

	return (
		<div className="container mx-auto p-6 space-y-6">
			{/* Controls */}
			<Card>
				<CardHeader>
					<CardTitle>Health Components Playground</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="flex flex-wrap gap-2">
						<Button
							variant={
								scenario === "healthy" ? "default" : "outline"
							}
							onClick={() => setScenario("healthy")}
						>
							Healthy
						</Button>
						<Button
							variant={
								scenario === "degraded" ? "default" : "outline"
							}
							onClick={() => setScenario("degraded")}
						>
							Degraded
						</Button>
						<Button
							variant={
								scenario === "unhealthy" ? "default" : "outline"
							}
							onClick={() => setScenario("unhealthy")}
						>
							Unhealthy
						</Button>
						<Button
							variant={
								scenario === "loading" ? "default" : "outline"
							}
							onClick={() => setScenario("loading")}
						>
							Loading
						</Button>
						<Button
							variant={
								scenario === "error" ? "default" : "outline"
							}
							onClick={() => setScenario("error")}
						>
							Error
						</Button>
					</div>
				</CardContent>
			</Card>

			{/* Auto Refresh Indicators */}
			<div className="space-y-4">
				<h2 className="text-xl font-semibold">
					Auto Refresh Indicator
				</h2>

				<div>
					<p className="text-sm text-muted-foreground mb-2">
						Versao Completa
					</p>
					<AutoRefreshIndicator
						interval={30000}
						lastRefresh={lastRefresh}
						onManualRefresh={handleRefresh}
						isRefreshing={isRefreshing}
					/>
				</div>

				<div>
					<p className="text-sm text-muted-foreground mb-2">
						Versao Compacta
					</p>
					<AutoRefreshIndicatorCompact
						interval={30000}
						lastRefresh={lastRefresh}
						onManualRefresh={handleRefresh}
						isRefreshing={isRefreshing}
					/>
				</div>
			</div>

			{/* Main Cards */}
			<div className="grid gap-6 md:grid-cols-2">
				<div className="space-y-2">
					<h2 className="text-xl font-semibold">
						Health Status Card
					</h2>
					<HealthStatusCard
						health={mockData.health}
						isLoading={scenario === "loading"}
						error={mockData.healthError}
						lastChecked={lastRefresh}
						onRetry={handleRefresh}
					/>
				</div>

				<div className="space-y-2">
					<h2 className="text-xl font-semibold">Readiness Card</h2>
					<ReadinessCard
						readiness={mockData.readiness}
						isLoading={scenario === "loading"}
						error={mockData.readinessError}
					/>
				</div>
			</div>

			{/* Individual Dependency Status */}
			<div className="space-y-4">
				<h2 className="text-xl font-semibold">
					Individual Dependency Status
				</h2>

				<Card>
					<CardHeader>
						<CardTitle className="text-base">
							Todos os Estados
						</CardTitle>
					</CardHeader>
					<CardContent className="space-y-0">
						<DependencyStatus
							name="database"
							status={{
								status: "healthy",
								duration_ms: 12,
								circuit_state: "closed",
							}}
						/>
						<DependencyStatus
							name="redis"
							status={{
								status: "degraded",
								duration_ms: 156,
								circuit_state: "half-open",
								error: "Latencia alta detectada",
							}}
						/>
						<DependencyStatus
							name="storage"
							status={{
								status: "unhealthy",
								duration_ms: 5000,
								circuit_state: "open",
								error: "Connection timeout after 5s",
								metadata: {
									last_successful: "2025-12-18T10:20:00Z",
									retry_count: 3,
								},
							}}
						/>
						<DependencyStatus
							name="whatsapp"
							status={{
								status: "healthy",
								duration_ms: 45,
							}}
						/>
					</CardContent>
				</Card>
			</div>

			{/* Metrics Display */}
			<div className="space-y-2">
				<h2 className="text-xl font-semibold">
					Metrics Display (Placeholder)
				</h2>
				<MetricsDisplay />
			</div>
		</div>
	);
}

function getMockData(scenario: string) {
	const baseHealth: HealthResponse = {
		status: "ok",
		service: "whatsapp-api",
		timestamp: new Date().toISOString(),
	};

	const baseReadiness: ReadinessResponse = {
		ready: true,
		observed_at: new Date().toISOString(),
		checks: {
			database: {
				status: "healthy",
				duration_ms: 12,
				circuit_state: "closed",
			},
			redis: {
				status: "healthy",
				duration_ms: 8,
				circuit_state: "closed",
			},
			storage: {
				status: "healthy",
				duration_ms: 45,
			},
		},
	};

	switch (scenario) {
		case "healthy":
			return {
				health: baseHealth,
				readiness: baseReadiness,
				healthError: undefined,
				readinessError: undefined,
			};

		case "degraded":
			return {
				health: baseHealth,
				readiness: {
					...baseReadiness,
					ready: true,
					checks: {
						...baseReadiness.checks,
						redis: {
							status: "degraded" as const,
							duration_ms: 156,
							circuit_state: "half-open" as const,
							error: "Latencia alta detectada",
						},
					},
				},
				healthError: undefined,
				readinessError: undefined,
			};

		case "unhealthy":
			return {
				health: baseHealth,
				readiness: {
					...baseReadiness,
					ready: false,
					checks: {
						database: {
							status: "unhealthy" as const,
							duration_ms: 5000,
							circuit_state: "open" as const,
							error: "Connection timeout after 5s",
						},
						redis: {
							status: "degraded" as const,
							duration_ms: 156,
							circuit_state: "half-open" as const,
							error: "Latencia alta detectada",
						},
						storage: {
							status: "healthy" as const,
							duration_ms: 45,
						},
					},
				},
				healthError: undefined,
				readinessError: undefined,
			};

		case "loading":
			return {
				health: undefined,
				readiness: undefined,
				healthError: undefined,
				readinessError: undefined,
			};

		case "error":
			return {
				health: undefined,
				readiness: undefined,
				healthError: new Error(
					"Failed to fetch health status: Network error",
				),
				readinessError: new Error(
					"Failed to fetch readiness status: Timeout",
				),
			};

		default:
			return {
				health: baseHealth,
				readiness: baseReadiness,
				healthError: undefined,
				readinessError: undefined,
			};
	}
}
