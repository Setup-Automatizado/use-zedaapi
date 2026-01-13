"use client";

import {
	ArrowLeft,
	BarChart3,
	Check,
	Copy,
	Key,
	Settings,
	TestTube,
	Webhook,
} from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { use, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	InstanceOverview,
	InstanceSettingsForm,
	InstanceStatistics,
	MessageTestForm,
	SubscriptionManagement,
	TokenDisplay,
	WebhookConfigForm,
} from "@/components/instances";
import { PageHeader } from "@/components/shared/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useClientToken, useInstance, useInstanceStatus } from "@/hooks";
import type { DeviceInfo } from "@/types";

interface InstancePageProps {
	params: Promise<{ id: string }>;
}

export default function InstancePage({ params }: InstancePageProps) {
	const resolvedParams = use(params);
	const router = useRouter();
	const searchParams = useSearchParams();
	const tabParam = searchParams.get("tab");
	const validTabs = [
		"overview",
		"statistics",
		"tokens",
		"test",
		"webhooks",
		"settings",
	];
	const defaultTab =
		tabParam && validTabs.includes(tabParam) ? tabParam : "overview";
	const { instance, isLoading, error, mutate } = useInstance(resolvedParams.id);
	const { isConnected, smartphoneConnected } = useInstanceStatus(
		resolvedParams.id,
		{
			enabled: true,
			interval: 5000,
		},
	);
	const { clientToken } = useClientToken();

	const [deviceInfo, setDeviceInfo] = useState<DeviceInfo | undefined>();
	const [isCurlCopied, setIsCurlCopied] = useState(false);

	// Mask token for display (show only last 4 chars)
	const maskToken = (token: string | null | undefined) => {
		if (!token || token.length < 8) return "********";
		return `${"*".repeat(token.length - 4)}${token.slice(-4)}`;
	};

	// Copy cURL command to clipboard
	const handleCopyCurl = async () => {
		const curlCommand = `curl -X POST \\
  ${process.env.NEXT_PUBLIC_WHATSAPP_API_URL}/instances/${instance?.id}/token/${instance?.instanceToken}/send-text \\
  -H "Content-Type: application/json" \\
  -H "Client-Token: ${clientToken}" \\
  -d '{
    "phone": "5511999999999",
    "message": "Hello from WhatsApp API!"
  }'`;

		try {
			await navigator.clipboard.writeText(curlCommand);
			setIsCurlCopied(true);
			toast.success("cURL command copied!", {
				description: "The command has been copied to your clipboard.",
			});
			setTimeout(() => setIsCurlCopied(false), 2000);
		} catch {
			toast.error("Failed to copy", {
				description: "Could not copy the cURL command.",
			});
		}
	};

	// Determine if we should fetch device info
	const shouldFetchDeviceInfo = Boolean(
		instance && isConnected && smartphoneConnected,
	);

	// Fetch device info when connected
	useEffect(() => {
		if (!shouldFetchDeviceInfo || !instance) {
			return;
		}

		let cancelled = false;

		const fetchDeviceInfo = async () => {
			try {
				const response = await fetch(`/api/instances/${instance.id}/device`, {
					method: "GET",
					headers: {
						"Content-Type": "application/json",
					},
				});

				if (response.ok && !cancelled) {
					const data = await response.json();
					setDeviceInfo(data);
				} else if (!cancelled) {
					setDeviceInfo(undefined);
				}
			} catch (err) {
				console.error("Failed to fetch device info:", err);
				if (!cancelled) {
					setDeviceInfo(undefined);
				}
			}
		};

		fetchDeviceInfo();

		return () => {
			cancelled = true;
		};
	}, [shouldFetchDeviceInfo, instance]);

	if (error) {
		return (
			<div className="space-y-6">
				<Button variant="ghost" onClick={() => router.back()} className="mb-4">
					<ArrowLeft className="mr-2 h-4 w-4" />
					Back
				</Button>
				<Alert variant="destructive">
					<AlertTitle>Error loading instance</AlertTitle>
					<AlertDescription>
						{error.message || "Could not load instance information."}
					</AlertDescription>
				</Alert>
			</div>
		);
	}

	if (isLoading || !instance) {
		return (
			<div className="space-y-6">
				<Skeleton className="h-8 w-64" />
				<div className="space-y-4">
					<Skeleton className="h-64 w-full" />
					<Skeleton className="h-96 w-full" />
				</div>
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<div className="flex items-center gap-4">
				<Button variant="ghost" onClick={() => router.back()} size="icon">
					<ArrowLeft className="h-4 w-4" />
				</Button>
				<PageHeader
					title={instance.name}
					description="Manage your WhatsApp instance"
				/>
			</div>

			<Tabs defaultValue={defaultTab} className="w-full">
				<TabsList>
					<TabsTrigger value="overview">Overview</TabsTrigger>
					<TabsTrigger value="statistics">
						<BarChart3 className="mr-2 h-4 w-4" />
						Statistics
					</TabsTrigger>
					<TabsTrigger value="tokens">
						<Key className="mr-2 h-4 w-4" />
						Tokens
					</TabsTrigger>
					<TabsTrigger value="test">
						<TestTube className="mr-2 h-4 w-4" />
						Test
					</TabsTrigger>
					<TabsTrigger value="webhooks">
						<Webhook className="mr-2 h-4 w-4" />
						Webhooks
					</TabsTrigger>
					<TabsTrigger value="settings">
						<Settings className="mr-2 h-4 w-4" />
						Settings
					</TabsTrigger>
				</TabsList>

				<TabsContent value="overview" className="mt-6">
					<div className="space-y-6">
						<InstanceOverview
							instance={instance}
							deviceInfo={deviceInfo}
							isConnected={isConnected}
							smartphoneConnected={smartphoneConnected}
						/>

						<SubscriptionManagement instance={instance} onUpdate={mutate} />
					</div>
				</TabsContent>

				<TabsContent value="statistics" className="mt-6">
					<InstanceStatistics instance={instance} />
				</TabsContent>

				<TabsContent value="tokens" className="mt-6">
					<div className="space-y-6">
						<Card>
							<CardHeader>
								<CardTitle>Authentication Tokens</CardTitle>
								<CardDescription>
									Use these tokens to authenticate API requests for this
									instance.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-6">
								<TokenDisplay
									label="Instance Token"
									token={instance.token}
									description="Instance-specific authentication token"
								/>

								<TokenDisplay
									label="Client Token"
									token={clientToken}
									description="Global client authentication token (same for all instances)"
								/>
							</CardContent>
						</Card>

						<Card>
							<CardHeader className="flex flex-row items-center justify-between">
								<div>
									<CardTitle>Usage Example</CardTitle>
									<CardDescription>
										Example cURL request to send a text message
									</CardDescription>
								</div>
								<Button
									variant="outline"
									size="icon"
									onClick={handleCopyCurl}
									title="Copy cURL command"
								>
									{isCurlCopied ? (
										<Check className="h-4 w-4 text-green-600" />
									) : (
										<Copy className="h-4 w-4" />
									)}
								</Button>
							</CardHeader>
							<CardContent>
								<p className="text-xs text-muted-foreground mb-2">
									Tokens are masked for security. Click copy to get the real
									values.
								</p>
								<pre className="rounded-lg bg-muted p-4 text-sm overflow-x-auto">
									<code>{`curl -X POST \\
  ${process.env.NEXT_PUBLIC_WHATSAPP_API_URL}/instances/${instance.id}/token/${maskToken(instance.instanceToken)}/send-text \\
  -H "Content-Type: application/json" \\
  -H "Client-Token: ${maskToken(clientToken)}" \\
  -d '{
    "phone": "5511999999999",
    "message": "Hello from WhatsApp API!"
  }'`}</code>
								</pre>
							</CardContent>
						</Card>
					</div>
				</TabsContent>

				<TabsContent value="test" className="mt-6">
					<MessageTestForm
						instanceId={instance.id}
						instanceToken={instance.instanceToken}
					/>
				</TabsContent>

				<TabsContent value="webhooks" className="mt-6">
					<Card>
						<CardHeader>
							<CardTitle>Webhook URLs</CardTitle>
							<CardDescription>
								Configure the endpoints to receive WhatsApp event notifications.
								Leave blank to disable specific webhooks.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<WebhookConfigForm
								instanceId={instance.id}
								instanceToken={instance.instanceToken}
								initialValues={{
									deliveryCallbackUrl: instance.deliveryCallbackUrl || "",
									receivedCallbackUrl: instance.receivedCallbackUrl || "",
									receivedAndDeliveryCallbackUrl:
										instance.receivedAndDeliveryCallbackUrl || "",
									messageStatusCallbackUrl:
										instance.messageStatusCallbackUrl || "",
									connectedCallbackUrl: instance.connectedCallbackUrl || "",
									disconnectedCallbackUrl:
										instance.disconnectedCallbackUrl || "",
									presenceChatCallbackUrl:
										instance.presenceChatCallbackUrl || "",
									notifySentByMe: instance.notifySentByMe || false,
								}}
							/>
						</CardContent>
					</Card>
				</TabsContent>

				<TabsContent value="settings" className="mt-6">
					<Card>
						<CardHeader>
							<CardTitle>Message and Call Settings</CardTitle>
							<CardDescription>
								Customize how your instance handles incoming messages and calls.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<InstanceSettingsForm
								instanceId={instance.id}
								instanceToken={instance.instanceToken}
								initialValues={{
									autoReadMessage: instance.autoReadMessage || false,
									callRejectAuto: instance.callRejectAuto || false,
									callRejectMessage: instance.callRejectMessage || "",
									notifySentByMe: instance.notifySentByMe || false,
								}}
							/>
						</CardContent>
					</Card>
				</TabsContent>
			</Tabs>
		</div>
	);
}
