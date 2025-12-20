"use client";

import { AlertCircle, Loader2, Power, RefreshCw, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

import { deleteInstance, disconnectInstance, restartInstance } from "@/actions";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { DeviceInfo, Instance } from "@/types";
import { isError } from "@/types";
import { DeviceInfoCard } from "./device-info-card";
import { InstanceStatusBadge } from "./instance-status-badge";
import { PhonePairingForm } from "./phone-pairing-form";
import { QRCodeDisplay } from "./qr-code-display";

export interface InstanceOverviewProps {
	instance: Instance;
	deviceInfo?: DeviceInfo;
	isConnected: boolean;
	smartphoneConnected: boolean;
}

export function InstanceOverview({
	instance,
	deviceInfo,
	isConnected,
	smartphoneConnected,
}: InstanceOverviewProps) {
	const router = useRouter();
	const [isRestarting, setIsRestarting] = useState(false);
	const [isDisconnecting, setIsDisconnecting] = useState(false);
	const [isDeleting, setIsDeleting] = useState(false);
	const [showDeleteDialog, setShowDeleteDialog] = useState(false);

	const handleRestart = async () => {
		setIsRestarting(true);
		try {
			const result = await restartInstance(instance.id, instance.token);

			if (isError(result)) {
				toast.error(result.error || "Error restarting instance");
				return;
			}

			toast.success("Instance restarted successfully!");
		} catch {
			toast.error("Error restarting instance");
		} finally {
			setIsRestarting(false);
		}
	};

	const handleDisconnect = async () => {
		setIsDisconnecting(true);
		try {
			const result = await disconnectInstance(instance.id, instance.token);

			if (isError(result)) {
				toast.error(result.error || "Error disconnecting instance");
				return;
			}

			toast.success("Instance disconnected successfully!");
		} catch {
			toast.error("Error disconnecting instance");
		} finally {
			setIsDisconnecting(false);
		}
	};

	const handleDelete = async () => {
		setIsDeleting(true);
		try {
			const result = await deleteInstance(instance.id);

			if (isError(result)) {
				toast.error(result.error || "Error deleting instance");
				setIsDeleting(false);
				return;
			}

			toast.success("Instance deleted successfully!");
			router.push("/instances");
		} catch {
			toast.error("Error deleting instance");
			setIsDeleting(false);
		}
	};

	const showConnectionOptions = !isConnected || !smartphoneConnected;

	return (
		<div className="space-y-6">
			{/* Status and Actions */}
			<Card>
				<CardHeader>
					<div className="flex items-start justify-between">
						<div className="space-y-1">
							<CardTitle>{instance.name}</CardTitle>
							<CardDescription>
								{instance.sessionName && `Session: ${instance.sessionName}`}
							</CardDescription>
						</div>
						<InstanceStatusBadge
							connected={isConnected}
							smartphoneConnected={smartphoneConnected}
						/>
					</div>
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Instance Details */}
					<div className="grid gap-3 text-sm">
						<div className="flex justify-between">
							<span className="text-muted-foreground">Instance ID:</span>
							<span className="font-mono text-xs">{instance.id}</span>
						</div>
						<div className="flex justify-between">
							<span className="text-muted-foreground">Middleware Type:</span>
							<span className="capitalize">{instance.middleware}</span>
						</div>
						{instance.created && (
							<div className="flex justify-between">
								<span className="text-muted-foreground">Created Date:</span>
								<span>
									{(() => {
										const createdDate = new Date(instance.created);
										return isNaN(createdDate.getTime())
											? "N/A"
											: createdDate.toLocaleDateString("en-US");
									})()}
								</span>
							</div>
						)}
						{instance.due && (
							<div className="flex justify-between">
								<span className="text-muted-foreground">
									Subscription Expiry:
								</span>
								<span
									className={
										instance.subscriptionActive
											? "text-green-600 dark:text-green-400"
											: "text-red-600 dark:text-red-400"
									}
								>
									{(() => {
										// due is Unix timestamp in milliseconds
										const dueDate = new Date(instance.due);
										return isNaN(dueDate.getTime())
											? "N/A"
											: dueDate.toLocaleDateString("en-US");
									})()}
								</span>
							</div>
						)}
					</div>

					{!instance.subscriptionActive && (
						<Alert variant="destructive">
							<AlertCircle className="h-4 w-4" />
							<AlertDescription>
								Subscription expired. Renew to continue using the instance.
							</AlertDescription>
						</Alert>
					)}

					<Separator />

					{/* Action Buttons */}
					<div className="flex flex-wrap gap-2">
						<Button
							onClick={handleRestart}
							disabled={isRestarting || isDisconnecting || isDeleting}
							variant="outline"
							size="sm"
						>
							{isRestarting ? (
								<Loader2 className="mr-2 h-4 w-4 animate-spin" />
							) : (
								<RefreshCw className="mr-2 h-4 w-4" />
							)}
							Restart
						</Button>

						{isConnected && (
							<Button
								onClick={handleDisconnect}
								disabled={isRestarting || isDisconnecting || isDeleting}
								variant="outline"
								size="sm"
							>
								{isDisconnecting ? (
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
								) : (
									<Power className="mr-2 h-4 w-4" />
								)}
								Disconnect
							</Button>
						)}

						<Button
							onClick={() => setShowDeleteDialog(true)}
							disabled={isRestarting || isDisconnecting || isDeleting}
							variant="destructive"
							size="sm"
						>
							{isDeleting ? (
								<Loader2 className="mr-2 h-4 w-4 animate-spin" />
							) : (
								<Trash2 className="mr-2 h-4 w-4" />
							)}
							Delete
						</Button>
					</div>
				</CardContent>
			</Card>

			{/* Device Info or Connection Options */}
			{deviceInfo && isConnected && smartphoneConnected ? (
				<DeviceInfoCard deviceInfo={deviceInfo} />
			) : showConnectionOptions ? (
				<Tabs defaultValue="qr-code" className="w-full">
					<TabsList className="grid w-full grid-cols-2">
						<TabsTrigger value="qr-code">QR Code</TabsTrigger>
						<TabsTrigger value="phone-pairing">Phone Pairing</TabsTrigger>
					</TabsList>
					<TabsContent value="qr-code" className="mt-4">
						<QRCodeDisplay
							instanceId={instance.id}
							instanceToken={instance.token}
						/>
					</TabsContent>
					<TabsContent value="phone-pairing" className="mt-4">
						<PhonePairingForm
							instanceId={instance.id}
							instanceToken={instance.token}
						/>
					</TabsContent>
				</Tabs>
			) : null}

			{/* Delete Confirmation Dialog */}
			<ConfirmDialog
				open={showDeleteDialog}
				onOpenChange={setShowDeleteDialog}
				onConfirm={handleDelete}
				title="Delete Instance"
				description={`Are you sure you want to delete the instance "${instance.name}"? This action cannot be undone.`}
				confirmLabel="Delete"
				cancelLabel="Cancel"
				variant="destructive"
			/>
		</div>
	);
}
