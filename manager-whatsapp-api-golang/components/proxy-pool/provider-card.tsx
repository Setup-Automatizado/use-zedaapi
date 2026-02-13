"use client";

import { Clock, MoreVertical, RefreshCw, Trash2 } from "lucide-react";
import { useTransition } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Switch } from "@/components/ui/switch";
import type { PoolProvider } from "@/types/pool";

interface ProviderCardProps {
	provider: PoolProvider;
	onSync: (id: string) => Promise<void>;
	onToggle: (id: string, enabled: boolean) => Promise<void>;
	onDelete: (id: string) => Promise<void>;
}

export function ProviderCard({
	provider,
	onSync,
	onToggle,
	onDelete,
}: ProviderCardProps) {
	const [isSyncing, startSync] = useTransition();
	const [isToggling, startToggle] = useTransition();
	const [isDeleting, startDelete] = useTransition();

	const handleSync = () => {
		startSync(async () => {
			try {
				await onSync(provider.id);
				toast.success("Sync started", {
					description: `Syncing ${provider.name}...`,
				});
			} catch {
				toast.error("Sync failed");
			}
		});
	};

	const handleToggle = (checked: boolean) => {
		startToggle(async () => {
			try {
				await onToggle(provider.id, checked);
				toast.success(
					checked ? "Provider enabled" : "Provider disabled",
				);
			} catch {
				toast.error("Failed to update provider");
			}
		});
	};

	const handleDelete = () => {
		startDelete(async () => {
			try {
				await onDelete(provider.id);
				toast.success("Provider deleted");
			} catch {
				toast.error("Failed to delete provider");
			}
		});
	};

	const formatTime = (dateStr?: string) => {
		if (!dateStr) return "Never";
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		if (diffMins < 1) return "Just now";
		if (diffMins < 60) return `${diffMins}m ago`;
		const diffHours = Math.floor(diffMins / 60);
		if (diffHours < 24) return `${diffHours}h ago`;
		return `${Math.floor(diffHours / 24)}d ago`;
	};

	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between pb-3">
				<div className="flex items-center gap-3">
					<CardTitle className="text-base">{provider.name}</CardTitle>
					<Badge variant={provider.enabled ? "default" : "secondary"}>
						{provider.enabled ? "Active" : "Disabled"}
					</Badge>
					<Badge variant="outline">{provider.providerType}</Badge>
				</div>
				<div className="flex items-center gap-2">
					<Switch
						checked={provider.enabled}
						onCheckedChange={handleToggle}
						disabled={isToggling}
					/>
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button
								variant="ghost"
								size="icon"
								className="h-8 w-8"
							>
								<MoreVertical className="h-4 w-4" />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								onClick={handleSync}
								disabled={isSyncing}
							>
								<RefreshCw className="mr-2 h-4 w-4" />
								Sync Now
							</DropdownMenuItem>
							<DropdownMenuItem
								onClick={handleDelete}
								disabled={isDeleting}
								className="text-destructive"
							>
								<Trash2 className="mr-2 h-4 w-4" />
								Delete
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
			</CardHeader>
			<CardContent>
				<div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
					<div>
						<p className="text-muted-foreground">Proxies</p>
						<p className="font-medium">{provider.proxyCount}</p>
					</div>
					<div>
						<p className="text-muted-foreground">Priority</p>
						<p className="font-medium">{provider.priority}</p>
					</div>
					<div>
						<p className="text-muted-foreground">Max per Proxy</p>
						<p className="font-medium">
							{provider.maxInstancesPerProxy}
						</p>
					</div>
					<div className="flex items-center gap-1">
						<Clock className="h-3 w-3 text-muted-foreground" />
						<p className="text-muted-foreground">Last Sync:</p>
						<p className="font-medium">
							{formatTime(provider.lastSyncAt)}
						</p>
					</div>
				</div>
				{provider.syncError && (
					<div className="mt-3 p-2 rounded bg-destructive/10 text-destructive text-xs">
						Sync error: {provider.syncError}
					</div>
				)}
				{provider.countryCodes.length > 0 && (
					<div className="mt-3 flex flex-wrap gap-1">
						{provider.countryCodes.map((code) => (
							<Badge
								key={code}
								variant="outline"
								className="text-xs"
							>
								{code}
							</Badge>
						))}
					</div>
				)}
			</CardContent>
		</Card>
	);
}
