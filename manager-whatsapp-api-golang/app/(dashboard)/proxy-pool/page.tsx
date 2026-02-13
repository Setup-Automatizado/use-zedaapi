"use client";

import { Globe, Loader2, Plus, RefreshCw, Shuffle, Users } from "lucide-react";
import { useCallback, useEffect, useState, useTransition } from "react";
import { toast } from "sonner";
import {
	bulkAssignPoolProxies,
	createPoolGroup,
	createPoolProvider,
	deletePoolGroup,
	deletePoolProvider,
	fetchPoolGroups,
	fetchPoolProviders,
	fetchPoolProxies,
	fetchPoolStats,
	syncPoolProvider,
	updatePoolProvider,
} from "@/actions/pool";
import {
	GroupCard,
	GroupForm,
	PoolStatsCards,
	ProviderCard,
	ProviderForm,
	ProxyTable,
} from "@/components/proxy-pool";
import { PageHeader } from "@/components/shared/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type {
	CreateGroupFormValues,
	CreateProviderFormValues,
} from "@/schemas/pool";
import type {
	BulkAssignResult,
	PoolGroup,
	PoolProvider as PoolProviderType,
	PoolProxy,
	PoolStats,
} from "@/types/pool";

export default function ProxyPoolPage() {
	const [stats, setStats] = useState<PoolStats | null>(null);
	const [providers, setProviders] = useState<PoolProviderType[]>([]);
	const [proxies, setProxies] = useState<PoolProxy[]>([]);
	const [proxyTotal, setProxyTotal] = useState(0);
	const [groups, setGroups] = useState<PoolGroup[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [showProviderForm, setShowProviderForm] = useState(false);
	const [showGroupForm, setShowGroupForm] = useState(false);
	const [bulkResult, setBulkResult] = useState<BulkAssignResult | null>(null);
	const [isRefreshing, startRefresh] = useTransition();
	const [isBulkAssigning, startBulkAssign] = useTransition();

	const loadData = useCallback(async () => {
		try {
			const [statsResult, providersResult, proxiesResult, groupsResult] =
				await Promise.all([
					fetchPoolStats(),
					fetchPoolProviders(),
					fetchPoolProxies({ limit: 50 }),
					fetchPoolGroups(),
				]);

			if (statsResult.success && statsResult.data)
				setStats(statsResult.data);
			if (providersResult.success && providersResult.data)
				setProviders(providersResult.data);
			if (proxiesResult.success && proxiesResult.data) {
				setProxies(proxiesResult.data.data ?? []);
				setProxyTotal(proxiesResult.data.total ?? 0);
			}
			if (groupsResult.success && groupsResult.data)
				setGroups(groupsResult.data);
		} catch {
			toast.error("Failed to load pool data");
		} finally {
			setIsLoading(false);
		}
	}, []);

	useEffect(() => {
		loadData();
	}, [loadData]);

	const handleRefresh = () => {
		startRefresh(async () => {
			setIsLoading(true);
			await loadData();
			toast.success("Data refreshed");
		});
	};

	const handleCreateProvider = async (data: CreateProviderFormValues) => {
		const result = await createPoolProvider(data);
		if (!result.success) throw new Error(result.error || "Failed");
		await loadData();
	};

	const handleSyncProvider = async (id: string) => {
		const result = await syncPoolProvider(id);
		if (!result.success) throw new Error(result.error || "Failed");
		setTimeout(() => loadData(), 2000);
	};

	const handleToggleProvider = async (id: string, enabled: boolean) => {
		const result = await updatePoolProvider(id, { enabled });
		if (!result.success) throw new Error(result.error || "Failed");
		await loadData();
	};

	const handleDeleteProvider = async (id: string) => {
		const result = await deletePoolProvider(id);
		if (!result.success) throw new Error(result.error || "Failed");
		await loadData();
	};

	const handleCreateGroup = async (data: CreateGroupFormValues) => {
		const result = await createPoolGroup(data);
		if (!result.success) throw new Error(result.error || "Failed");
		await loadData();
	};

	const handleDeleteGroup = async (id: string) => {
		const result = await deletePoolGroup(id);
		if (!result.success) throw new Error(result.error || "Failed");
		await loadData();
	};

	const handleBulkAssign = () => {
		if (
			!confirm(
				"This will assign pool proxies to ALL unassigned active instances. Continue?",
			)
		)
			return;
		startBulkAssign(async () => {
			setBulkResult(null);
			const result = await bulkAssignPoolProxies({});
			if (result.success && result.data) {
				setBulkResult(result.data);
				toast.success(
					`Bulk assign complete: ${result.data.assigned} assigned, ${result.data.skipped} skipped, ${result.data.failed} failed`,
				);
				await loadData();
			} else {
				toast.error(result.error || "Bulk assign failed");
			}
		});
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<PageHeader
					title="Proxy Pool"
					description="Manage proxy providers, pool, and group assignments"
				/>
				<div className="flex items-center gap-2">
					<Button
						size="sm"
						onClick={handleBulkAssign}
						disabled={isBulkAssigning || proxyTotal === 0}
					>
						{isBulkAssigning ? (
							<Loader2 className="mr-2 h-4 w-4 animate-spin" />
						) : (
							<Shuffle className="mr-2 h-4 w-4" />
						)}
						{isBulkAssigning ? "Assigning..." : "Auto-Assign All"}
					</Button>
					<Button
						variant="outline"
						size="sm"
						onClick={handleRefresh}
						disabled={isRefreshing}
					>
						<RefreshCw
							className={`mr-2 h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
						/>
						Refresh
					</Button>
				</div>
			</div>

			{bulkResult && (
				<Alert
					variant={bulkResult.failed > 0 ? "destructive" : "default"}
				>
					<AlertTitle>Bulk Assignment Result</AlertTitle>
					<AlertDescription>
						<div className="flex gap-4 text-sm mt-1">
							<span>
								Total: <strong>{bulkResult.total}</strong>
							</span>
							<span>
								Assigned: <strong>{bulkResult.assigned}</strong>
							</span>
							<span>
								Skipped: <strong>{bulkResult.skipped}</strong>
							</span>
							{bulkResult.failed > 0 && (
								<span>
									Failed: <strong>{bulkResult.failed}</strong>
								</span>
							)}
						</div>
						{bulkResult.errors && bulkResult.errors.length > 0 && (
							<details className="mt-2">
								<summary className="text-xs cursor-pointer text-muted-foreground">
									View errors ({bulkResult.errors.length})
								</summary>
								<ul className="text-xs mt-1 space-y-0.5 text-muted-foreground">
									{bulkResult.errors.map((e, i) => (
										<li key={i}>{e}</li>
									))}
								</ul>
							</details>
						)}
					</AlertDescription>
				</Alert>
			)}

			<PoolStatsCards stats={stats} isLoading={isLoading} />

			<Tabs defaultValue="providers" className="w-full">
				<TabsList>
					<TabsTrigger value="providers">
						<Globe className="mr-2 h-4 w-4" />
						Providers
					</TabsTrigger>
					<TabsTrigger value="proxies">
						Proxies ({proxyTotal})
					</TabsTrigger>
					<TabsTrigger value="groups">
						<Users className="mr-2 h-4 w-4" />
						Groups
					</TabsTrigger>
				</TabsList>

				<TabsContent value="providers" className="mt-6 space-y-4">
					<div className="flex justify-end">
						<Button onClick={() => setShowProviderForm(true)}>
							<Plus className="mr-2 h-4 w-4" />
							Add Provider
						</Button>
					</div>
					{providers.length === 0 && !isLoading ? (
						<Card>
							<CardContent className="flex flex-col items-center justify-center py-12">
								<Globe className="h-12 w-12 text-muted-foreground mb-4" />
								<p className="text-muted-foreground">
									No providers configured
								</p>
								<p className="text-sm text-muted-foreground mt-1">
									Add a proxy provider to start sourcing
									proxies.
								</p>
								<Button
									className="mt-4"
									onClick={() => setShowProviderForm(true)}
								>
									<Plus className="mr-2 h-4 w-4" />
									Add Provider
								</Button>
							</CardContent>
						</Card>
					) : (
						<div className="space-y-4">
							{providers.map((provider) => (
								<ProviderCard
									key={provider.id}
									provider={provider}
									onSync={handleSyncProvider}
									onToggle={handleToggleProvider}
									onDelete={handleDeleteProvider}
								/>
							))}
						</div>
					)}
				</TabsContent>

				<TabsContent value="proxies" className="mt-6">
					<Card>
						<CardHeader>
							<CardTitle>Pool Proxies</CardTitle>
						</CardHeader>
						<CardContent>
							<ProxyTable
								proxies={proxies}
								total={proxyTotal}
								isLoading={isLoading}
							/>
						</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="groups" className="mt-6 space-y-4">
					<div className="flex justify-end">
						<Button onClick={() => setShowGroupForm(true)}>
							<Plus className="mr-2 h-4 w-4" />
							Create Group
						</Button>
					</div>
					{groups.length === 0 && !isLoading ? (
						<Card>
							<CardContent className="flex flex-col items-center justify-center py-12">
								<Users className="h-12 w-12 text-muted-foreground mb-4" />
								<p className="text-muted-foreground">
									No groups configured
								</p>
								<p className="text-sm text-muted-foreground mt-1">
									Create groups to share proxies across
									multiple instances.
								</p>
							</CardContent>
						</Card>
					) : (
						<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
							{groups.map((group) => (
								<GroupCard
									key={group.id}
									group={group}
									onDelete={handleDeleteGroup}
								/>
							))}
						</div>
					)}
				</TabsContent>
			</Tabs>

			<ProviderForm
				open={showProviderForm}
				onOpenChange={setShowProviderForm}
				onSubmit={handleCreateProvider}
			/>
			<GroupForm
				open={showGroupForm}
				onOpenChange={setShowGroupForm}
				onSubmit={handleCreateGroup}
			/>
		</div>
	);
}
