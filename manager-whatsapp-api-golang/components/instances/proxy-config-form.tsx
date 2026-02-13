"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import {
	Activity,
	AlertTriangle,
	ArrowRightLeft,
	CheckCircle2,
	Clock,
	Eye,
	EyeOff,
	Loader2,
	Save,
	TestTube,
	Trash2,
	XCircle,
	Zap,
} from "lucide-react";
import { useCallback, useEffect, useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
	fetchProxyConfig,
	fetchProxyHealth,
	removeProxyConfig,
	swapProxyConnection,
	testProxyConnection,
	updateProxyConfig,
} from "@/actions/proxy";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import {
	type ProxyConfigFormValues,
	type ProxyConfigInput,
	ProxyConfigSchema,
} from "@/schemas/proxy";
import type { ProxyConfig, ProxyHealthLog } from "@/types/proxy";
import { sanitizeProxyUrlForDisplay } from "@/types/proxy";

export interface ProxyConfigFormProps {
	instanceId: string;
	instanceToken: string;
	onSuccess?: () => void;
}

export function ProxyConfigForm({
	instanceId,
	instanceToken,
	onSuccess,
}: ProxyConfigFormProps) {
	const [isPending, startTransition] = useTransition();
	const [isTesting, setIsTesting] = useState(false);
	const [isSwapping, setIsSwapping] = useState(false);
	const [isRemoving, setIsRemoving] = useState(false);
	const [isLoadingConfig, setIsLoadingConfig] = useState(true);
	const [showProxyUrl, setShowProxyUrl] = useState(false);
	const [currentConfig, setCurrentConfig] = useState<ProxyConfig | null>(null);
	const [healthLogs, setHealthLogs] = useState<ProxyHealthLog[]>([]);
	const [testResult, setTestResult] = useState<{
		reachable: boolean;
		latencyMs?: number;
		error?: string;
	} | null>(null);

	const form = useForm<ProxyConfigInput, unknown, ProxyConfigFormValues>({
		resolver: zodResolver(ProxyConfigSchema),
		defaultValues: {
			proxyUrl: "",
			noWebsocket: false,
			onlyLogin: false,
			noMedia: false,
		},
	});

	const {
		formState: { isDirty },
		reset,
	} = form;

	// Load current proxy config
	const loadConfig = useCallback(async () => {
		setIsLoadingConfig(true);
		try {
			const result = await fetchProxyConfig(instanceId, instanceToken);
			if (result.success && result.data?.proxy) {
				const cfg = result.data.proxy;
				setCurrentConfig(cfg);
				reset({
					proxyUrl: cfg.proxyUrl || "",
					noWebsocket: cfg.noWebsocket || false,
					onlyLogin: cfg.onlyLogin || false,
					noMedia: cfg.noMedia || false,
				});
			}
		} catch {
			// No proxy configured yet
		} finally {
			setIsLoadingConfig(false);
		}
	}, [instanceId, instanceToken, reset]);

	// Load health logs
	const loadHealth = useCallback(async () => {
		try {
			const result = await fetchProxyHealth(instanceId, instanceToken);
			if (result.success && result.data) {
				setHealthLogs(result.data.logs || []);
				if (result.data.proxy) {
					setCurrentConfig(result.data.proxy);
				}
			}
		} catch {
			// Ignore health fetch errors
		}
	}, [instanceId, instanceToken]);

	useEffect(() => {
		loadConfig();
		loadHealth();
	}, [loadConfig, loadHealth]);

	// Save proxy configuration
	const onSubmit = async (data: ProxyConfigFormValues) => {
		startTransition(async () => {
			try {
				const result = await updateProxyConfig(instanceId, instanceToken, data);
				if (result.success) {
					toast.success("Proxy configuration saved");
					if (result.data?.proxy) {
						setCurrentConfig(result.data.proxy);
					}
					setShowProxyUrl(false);
					reset(data);
					onSuccess?.();
					loadHealth();
				} else {
					toast.error(result.error || "Failed to save proxy");
				}
			} catch {
				toast.error("Failed to save proxy configuration");
			}
		});
	};

	// Test proxy connectivity
	const handleTest = async () => {
		const proxyUrl = form.getValues("proxyUrl");
		if (!proxyUrl) {
			toast.error("Enter a proxy URL first");
			return;
		}

		setIsTesting(true);
		setTestResult(null);
		try {
			const result = await testProxyConnection(
				instanceId,
				instanceToken,
				proxyUrl,
			);
			if (result.success && result.data) {
				setTestResult(result.data);
				if (result.data.reachable) {
					toast.success(`Proxy reachable (${result.data.latencyMs}ms latency)`);
				} else {
					toast.error(`Proxy unreachable: ${result.data.error}`);
				}
			} else {
				toast.error(result.error || "Test failed");
			}
		} catch {
			toast.error("Failed to test proxy");
		} finally {
			setIsTesting(false);
		}
	};

	// Hot-swap proxy
	const handleSwap = async () => {
		const proxyUrl = form.getValues("proxyUrl");
		if (!proxyUrl) {
			toast.error("Enter a proxy URL first");
			return;
		}

		setIsSwapping(true);
		try {
			const result = await swapProxyConnection(
				instanceId,
				instanceToken,
				proxyUrl,
			);
			if (result.success) {
				toast.success("Proxy swapped successfully - client reconnected");
				if (result.data?.proxy) {
					setCurrentConfig(result.data.proxy);
				}
				setShowProxyUrl(false);
				onSuccess?.();
				loadHealth();
			} else {
				toast.error(result.error || "Failed to swap proxy");
			}
		} catch {
			toast.error("Failed to swap proxy");
		} finally {
			setIsSwapping(false);
		}
	};

	// Remove proxy
	const handleRemove = async () => {
		setIsRemoving(true);
		try {
			const result = await removeProxyConfig(instanceId, instanceToken);
			if (result.success) {
				toast.success("Proxy configuration removed");
				setCurrentConfig(null);
				setTestResult(null);
				setShowProxyUrl(false);
				reset({
					proxyUrl: "",
					noWebsocket: false,
					onlyLogin: false,
					noMedia: false,
				});
				onSuccess?.();
			} else {
				toast.error(result.error || "Failed to remove proxy");
			}
		} catch {
			toast.error("Failed to remove proxy");
		} finally {
			setIsRemoving(false);
		}
	};

	const isAnyPending = isPending || isTesting || isSwapping || isRemoving;

	if (isLoadingConfig) {
		return (
			<div className="flex items-center justify-center py-8">
				<Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* Health Status Banner */}
			{currentConfig && currentConfig.proxyEnabled && (
				<div
					className={`flex items-center gap-3 rounded-lg border p-4 ${
						currentConfig.healthStatus === "healthy"
							? "border-green-200 bg-green-50/50 dark:border-green-900 dark:bg-green-950/30"
							: currentConfig.healthStatus === "unhealthy"
								? "border-red-200 bg-red-50/50 dark:border-red-900 dark:bg-red-950/30"
								: "border-yellow-200 bg-yellow-50/50 dark:border-yellow-900 dark:bg-yellow-950/30"
					}`}
				>
					<HealthStatusIcon status={currentConfig.healthStatus} />
					<div className="flex-1 min-w-0">
						<div className="flex items-center gap-2">
							<p className="text-sm font-medium">Proxy Status</p>
							<HealthStatusBadge status={currentConfig.healthStatus} />
						</div>
						{currentConfig.proxyUrl && (
							<p className="text-xs text-muted-foreground font-mono truncate mt-0.5">
								{sanitizeProxyUrlForDisplay(currentConfig.proxyUrl)}
							</p>
						)}
						<div className="flex items-center gap-3 mt-1">
							{currentConfig.healthFailures > 0 && (
								<span className="text-xs text-muted-foreground">
									{currentConfig.healthFailures} consecutive failure
									{currentConfig.healthFailures !== 1 ? "s" : ""}
								</span>
							)}
							{currentConfig.lastHealthCheck && (
								<span className="flex items-center gap-1 text-xs text-muted-foreground">
									<Clock className="h-3 w-3" />
									{new Date(currentConfig.lastHealthCheck).toLocaleString()}
								</span>
							)}
						</div>
					</div>
				</div>
			)}

			<Form {...form}>
				<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
					{/* Proxy URL */}
					<FormField
						control={form.control}
						name="proxyUrl"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Proxy URL</FormLabel>
								<div className="flex gap-2">
									<div className="relative flex-1">
										<FormControl>
											<Input
												type={showProxyUrl ? "text" : "password"}
												placeholder="socks5://user:pass@host:port"
												autoComplete="off"
												className="pr-10"
												{...field}
												disabled={isAnyPending}
											/>
										</FormControl>
										<Button
											type="button"
											variant="ghost"
											size="icon"
											className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
											onClick={() => setShowProxyUrl(!showProxyUrl)}
											tabIndex={-1}
											title={showProxyUrl ? "Hide proxy URL" : "Show proxy URL"}
										>
											{showProxyUrl ? (
												<EyeOff className="h-4 w-4 text-muted-foreground" />
											) : (
												<Eye className="h-4 w-4 text-muted-foreground" />
											)}
										</Button>
									</div>
									<TooltipProvider>
										<Tooltip>
											<TooltipTrigger asChild>
												<Button
													type="button"
													variant="outline"
													size="icon"
													onClick={handleTest}
													disabled={isAnyPending || !field.value}
												>
													{isTesting ? (
														<Loader2 className="h-4 w-4 animate-spin" />
													) : (
														<TestTube className="h-4 w-4" />
													)}
												</Button>
											</TooltipTrigger>
											<TooltipContent>Test proxy connectivity</TooltipContent>
										</Tooltip>
									</TooltipProvider>
								</div>
								<FormDescription>
									Supported schemes: http://, https://, socks5://
								</FormDescription>
								<FormMessage />
							</FormItem>
						)}
					/>

					{/* Test Result */}
					{testResult && (
						<div
							className={`flex items-center gap-2 rounded-lg border p-3 text-sm ${
								testResult.reachable
									? "border-green-200 bg-green-50 text-green-800 dark:border-green-800 dark:bg-green-950 dark:text-green-200"
									: "border-red-200 bg-red-50 text-red-800 dark:border-red-800 dark:bg-red-950 dark:text-red-200"
							}`}
						>
							{testResult.reachable ? (
								<>
									<CheckCircle2 className="h-4 w-4 shrink-0" />
									<span className="flex-1">Proxy reachable</span>
									{testResult.latencyMs !== undefined && (
										<Badge
											variant="outline"
											className="border-green-300 text-green-700 dark:border-green-700 dark:text-green-300 ml-auto"
										>
											<Zap className="mr-1 h-3 w-3" />
											{testResult.latencyMs}ms
										</Badge>
									)}
								</>
							) : (
								<>
									<XCircle className="h-4 w-4 shrink-0" />
									<span>Unreachable: {testResult.error}</span>
								</>
							)}
						</div>
					)}

					<Separator />

					{/* Advanced Options */}
					<div className="space-y-4">
						<h4 className="text-sm font-medium">Advanced Options</h4>

						<FormField
							control={form.control}
							name="noWebsocket"
							render={({ field }) => (
								<FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
									<div className="space-y-0.5">
										<FormLabel className="text-sm">No WebSocket</FormLabel>
										<FormDescription>
											Disable WebSocket connections through the proxy
										</FormDescription>
									</div>
									<FormControl>
										<Switch
											checked={field.value}
											onCheckedChange={field.onChange}
											disabled={isAnyPending}
										/>
									</FormControl>
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="onlyLogin"
							render={({ field }) => (
								<FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
									<div className="space-y-0.5">
										<FormLabel className="text-sm">Only Login</FormLabel>
										<FormDescription>
											Use the proxy only during login/registration
										</FormDescription>
									</div>
									<FormControl>
										<Switch
											checked={field.value}
											onCheckedChange={field.onChange}
											disabled={isAnyPending}
										/>
									</FormControl>
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="noMedia"
							render={({ field }) => (
								<FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
									<div className="space-y-0.5">
										<FormLabel className="text-sm">No Media</FormLabel>
										<FormDescription>
											Disable media upload/download through the proxy
										</FormDescription>
									</div>
									<FormControl>
										<Switch
											checked={field.value}
											onCheckedChange={field.onChange}
											disabled={isAnyPending}
										/>
									</FormControl>
								</FormItem>
							)}
						/>
					</div>

					<Separator />

					{/* Action Buttons */}
					<div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-between">
						<div className="flex gap-2">
							{currentConfig?.proxyEnabled && (
								<Button
									type="button"
									variant="destructive"
									onClick={handleRemove}
									disabled={isAnyPending}
								>
									{isRemoving ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Removing...
										</>
									) : (
										<>
											<Trash2 className="mr-2 h-4 w-4" />
											Remove Proxy
										</>
									)}
								</Button>
							)}
						</div>

						<div className="flex gap-2">
							{currentConfig?.proxyEnabled && (
								<TooltipProvider>
									<Tooltip>
										<TooltipTrigger asChild>
											<Button
												type="button"
												variant="outline"
												onClick={handleSwap}
												disabled={isAnyPending || !form.getValues("proxyUrl")}
											>
												{isSwapping ? (
													<>
														<Loader2 className="mr-2 h-4 w-4 animate-spin" />
														Swapping...
													</>
												) : (
													<>
														<ArrowRightLeft className="mr-2 h-4 w-4" />
														Hot Swap
													</>
												)}
											</Button>
										</TooltipTrigger>
										<TooltipContent>
											Hot-swap proxy without losing WhatsApp session
										</TooltipContent>
									</Tooltip>
								</TooltipProvider>
							)}

							<Button
								type="button"
								variant="outline"
								onClick={() => {
									reset();
									setShowProxyUrl(false);
								}}
								disabled={isAnyPending || !isDirty}
							>
								Cancel
							</Button>
							<Button type="submit" disabled={isAnyPending || !isDirty}>
								{isPending ? (
									<>
										<Loader2 className="mr-2 h-4 w-4 animate-spin" />
										Saving...
									</>
								) : (
									<>
										<Save className="mr-2 h-4 w-4" />
										Save
									</>
								)}
							</Button>
						</div>
					</div>

					{isDirty && !isPending && (
						<p className="text-center text-sm text-muted-foreground">
							You have unsaved changes
						</p>
					)}
				</form>
			</Form>

			{/* Health Check Logs */}
			{healthLogs.length > 0 && (
				<>
					<Separator />
					<div className="space-y-3">
						<div className="flex items-center gap-2">
							<Activity className="h-4 w-4" />
							<h4 className="text-sm font-medium">Recent Health Checks</h4>
						</div>
						<div className="space-y-2">
							{healthLogs.slice(0, 10).map((log, i) => (
								<div
									key={`${log.checkedAt}-${i}`}
									className="flex items-center justify-between rounded-lg border p-3 text-sm"
								>
									<div className="flex items-center gap-2 min-w-0 flex-1">
										<HealthStatusIcon status={log.status} />
										<div className="min-w-0 flex-1">
											<span className="text-muted-foreground text-xs">
												{new Date(log.checkedAt).toLocaleString()}
											</span>
											{log.proxyUrl && (
												<p className="text-xs font-mono text-muted-foreground/70 truncate">
													{sanitizeProxyUrlForDisplay(log.proxyUrl)}
												</p>
											)}
											{log.errorMessage && (
												<p className="text-xs text-red-600 dark:text-red-400 truncate">
													{log.errorMessage}
												</p>
											)}
										</div>
									</div>
									<div className="flex items-center gap-3 shrink-0 ml-3">
										{log.latencyMs !== undefined && log.latencyMs > 0 && (
											<span className="text-xs text-muted-foreground tabular-nums">
												{log.latencyMs}ms
											</span>
										)}
										<HealthStatusBadge status={log.status} />
									</div>
								</div>
							))}
						</div>
					</div>
				</>
			)}
		</div>
	);
}

function HealthStatusIcon({ status }: { status: string }) {
	switch (status) {
		case "healthy":
			return <CheckCircle2 className="h-4 w-4 shrink-0 text-green-500" />;
		case "unhealthy":
			return <XCircle className="h-4 w-4 shrink-0 text-red-500" />;
		default:
			return <AlertTriangle className="h-4 w-4 shrink-0 text-yellow-500" />;
	}
}

function HealthStatusBadge({ status }: { status: string }) {
	switch (status) {
		case "healthy":
			return (
				<Badge
					variant="outline"
					className="border-green-300 text-green-700 dark:border-green-700 dark:text-green-300"
				>
					Healthy
				</Badge>
			);
		case "unhealthy":
			return <Badge variant="destructive">Unhealthy</Badge>;
		default:
			return (
				<Badge
					variant="outline"
					className="border-yellow-300 text-yellow-700 dark:border-yellow-700 dark:text-yellow-300"
				>
					Unknown
				</Badge>
			);
	}
}
