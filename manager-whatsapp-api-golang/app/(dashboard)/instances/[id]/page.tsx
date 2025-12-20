"use client";

import { use, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Settings, Webhook, ArrowLeft } from "lucide-react";

import { useInstance, useInstanceStatus } from "@/hooks";
import {
	InstanceOverview,
	WebhookConfigForm,
	InstanceSettingsForm,
} from "@/components/instances";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import type { DeviceInfo } from "@/types";

interface InstancePageProps {
	params: Promise<{ id: string }>;
}

export default function InstancePage({ params }: InstancePageProps) {
	const resolvedParams = use(params);
	const router = useRouter();
	const searchParams = useSearchParams();
	const tabParam = searchParams.get("tab");
	const validTabs = ["overview", "webhooks", "settings"];
	const defaultTab =
		tabParam && validTabs.includes(tabParam) ? tabParam : "overview";
	const { instance, isLoading, error } = useInstance(resolvedParams.id);
	const { isConnected, smartphoneConnected } = useInstanceStatus(
		resolvedParams.id,
		{
			enabled: true,
			interval: 5000,
		},
	);

	const [deviceInfo, setDeviceInfo] = useState<DeviceInfo | undefined>();

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
				const response = await fetch(
					`/api/instances/${instance.id}/device`,
					{
						method: "GET",
						headers: {
							"Content-Type": "application/json",
						},
					},
				);

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
				<Button
					variant="ghost"
					onClick={() => router.back()}
					className="mb-4"
				>
					<ArrowLeft className="mr-2 h-4 w-4" />
					Back
				</Button>
				<Alert variant="destructive">
					<AlertTitle>Error loading instance</AlertTitle>
					<AlertDescription>
						{error.message ||
							"Could not load instance information."}
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
				<Button
					variant="ghost"
					onClick={() => router.back()}
					size="icon"
				>
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
					<InstanceOverview
						instance={instance}
						deviceInfo={deviceInfo}
						isConnected={isConnected}
						smartphoneConnected={smartphoneConnected}
					/>
				</TabsContent>

				<TabsContent value="webhooks" className="mt-6">
					<Card>
						<CardHeader>
							<CardTitle>Webhook URLs</CardTitle>
							<CardDescription>
								Configure the endpoints to receive WhatsApp
								event notifications. Leave blank to disable
								specific webhooks.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<WebhookConfigForm
								instanceId={instance.id}
								instanceToken={instance.instanceToken}
								initialValues={{
									deliveryCallbackUrl:
										instance.deliveryCallbackUrl || "",
									receivedCallbackUrl:
										instance.receivedCallbackUrl || "",
									receivedAndDeliveryCallbackUrl:
										instance.receivedAndDeliveryCallbackUrl ||
										"",
									messageStatusCallbackUrl:
										instance.messageStatusCallbackUrl || "",
									connectedCallbackUrl:
										instance.connectedCallbackUrl || "",
									disconnectedCallbackUrl:
										instance.disconnectedCallbackUrl || "",
									presenceChatCallbackUrl:
										instance.presenceChatCallbackUrl || "",
									notifySentByMe:
										instance.notifySentByMe || false,
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
								Customize how your instance handles incoming
								messages and calls.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<InstanceSettingsForm
								instanceId={instance.id}
								instanceToken={instance.instanceToken}
								initialValues={{
									autoReadMessage:
										instance.autoReadMessage || false,
									callRejectAuto:
										instance.callRejectAuto || false,
									callRejectMessage:
										instance.callRejectMessage || "",
									notifySentByMe:
										instance.notifySentByMe || false,
								}}
							/>
						</CardContent>
					</Card>
				</TabsContent>
			</Tabs>
		</div>
	);
}
