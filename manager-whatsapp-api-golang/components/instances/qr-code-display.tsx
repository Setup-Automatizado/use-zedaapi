"use client";

import {
	AlertCircle,
	CheckCircle2,
	Loader2,
	QrCode,
	RefreshCw,
} from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useQRCode } from "@/hooks";
import { cn } from "@/lib/utils";

export interface QRCodeDisplayProps {
	instanceId: string;
	instanceToken: string;
	className?: string;
}

export function QRCodeDisplay({
	instanceId,
	// instanceToken is kept in the interface for API consistency
	instanceToken: _instanceToken,
	className,
}: QRCodeDisplayProps) {
	void _instanceToken; // Mark as intentionally unused
	const [showQRCode, setShowQRCode] = useState(false);

	const { image, isConnected, isLoading, error, isRefreshing, refresh } =
		useQRCode(instanceId, {
			interval: 30000,
			autoPoll: showQRCode,
			enabled: showQRCode,
		});

	const handleGenerateQRCode = () => {
		setShowQRCode(true);
	};

	// Show success message when connected
	if (isConnected) {
		return (
			<Card className={cn("border-green-200 dark:border-green-800", className)}>
				<CardContent className="flex flex-col items-center justify-center py-12">
					<div className="mb-4 rounded-full bg-green-100 p-3 dark:bg-green-950">
						<CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
					</div>
					<h3 className="mb-2 text-lg font-semibold text-green-900 dark:text-green-100">
						WhatsApp Connected!
					</h3>
					<p className="text-center text-sm text-green-700 dark:text-green-300">
						Your instance is connected and ready to use
					</p>
				</CardContent>
			</Card>
		);
	}

	// Initial state - waiting for user to request QR Code
	if (!showQRCode) {
		return (
			<Card className={className}>
				<CardHeader>
					<CardTitle>Connect via QR Code</CardTitle>
					<CardDescription>
						Generate a QR Code to connect your instance to WhatsApp
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="flex flex-col items-center justify-center py-8">
						<div className="mb-4 rounded-full bg-muted p-4">
							<QrCode className="h-12 w-12 text-muted-foreground" />
						</div>
						<p className="mb-4 text-center text-sm text-muted-foreground">
							Click the button below to generate a connection QR Code
						</p>
						<Button onClick={handleGenerateQRCode}>
							<QrCode className="mr-2 h-4 w-4" />
							Generate QR Code
						</Button>
					</div>

					<div className="rounded-md bg-muted p-4">
						<h4 className="mb-2 text-sm font-medium">How to connect:</h4>
						<ol className="list-inside list-decimal space-y-1 text-sm text-muted-foreground">
							<li>Click &quot;Generate QR Code&quot;</li>
							<li>Open WhatsApp on your phone</li>
							<li>Tap Menu (3 dots) and select &quot;Linked devices&quot;</li>
							<li>Tap &quot;Link a device&quot;</li>
							<li>Point your phone at this screen to scan the code</li>
						</ol>
					</div>
				</CardContent>
			</Card>
		);
	}

	// Show loading state
	if (isLoading && !image) {
		return (
			<Card className={className}>
				<CardContent className="flex flex-col items-center justify-center py-12">
					<Loader2 className="mb-4 h-12 w-12 animate-spin text-muted-foreground" />
					<p className="text-sm text-muted-foreground">Generating QR Code...</p>
				</CardContent>
			</Card>
		);
	}

	// Show error state
	if (error && !image) {
		return (
			<Card className={cn("border-destructive", className)}>
				<CardContent className="py-8">
					<Alert variant="destructive">
						<AlertCircle className="h-4 w-4" />
						<AlertDescription className="ml-2">
							{error || "Error loading QR Code"}
						</AlertDescription>
					</Alert>
					<div className="mt-4 flex justify-center gap-2">
						<Button onClick={refresh} variant="outline" size="sm">
							<RefreshCw className="mr-2 h-4 w-4" />
							Try Again
						</Button>
						<Button
							onClick={() => setShowQRCode(false)}
							variant="ghost"
							size="sm"
						>
							Cancel
						</Button>
					</div>
				</CardContent>
			</Card>
		);
	}

	// Show QR code
	return (
		<Card className={className}>
			<CardHeader>
				<CardTitle>Scan QR Code</CardTitle>
				<CardDescription>
					Open WhatsApp on your phone and scan this code to connect
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				{image ? (
					<div className="flex flex-col items-center gap-4">
						<div className="relative aspect-square w-full max-w-xs overflow-hidden rounded-lg border bg-white p-4">
							<Image
								src={image}
								alt="QR Code to connect WhatsApp"
								fill
								className="object-contain"
								priority
							/>
						</div>

						<div className="flex items-center gap-2 text-sm text-muted-foreground">
							{isRefreshing ? (
								<>
									<Loader2 className="h-4 w-4 animate-spin" />
									<span>Updating QR Code...</span>
								</>
							) : (
								<>
									<span>QR Code auto-refreshes every 30 seconds</span>
								</>
							)}
						</div>

						<div className="flex gap-2">
							<Button
								onClick={refresh}
								variant="outline"
								size="sm"
								disabled={isRefreshing}
							>
								<RefreshCw
									className={cn("mr-2 h-4 w-4", isRefreshing && "animate-spin")}
								/>
								Refresh Now
							</Button>
							<Button
								onClick={() => setShowQRCode(false)}
								variant="ghost"
								size="sm"
							>
								Cancel
							</Button>
						</div>
					</div>
				) : (
					<div className="flex flex-col items-center justify-center py-8">
						<Loader2 className="mb-4 h-8 w-8 animate-spin text-muted-foreground" />
						<p className="text-sm text-muted-foreground">
							Generating QR Code...
						</p>
					</div>
				)}

				<div className="rounded-md bg-muted p-4">
					<h4 className="mb-2 text-sm font-medium">How to connect:</h4>
					<ol className="list-inside list-decimal space-y-1 text-sm text-muted-foreground">
						<li>Open WhatsApp on your phone</li>
						<li>Tap Menu (3 dots) and select &quot;Linked devices&quot;</li>
						<li>Tap &quot;Link a device&quot;</li>
						<li>Point your phone at this screen to scan the code</li>
					</ol>
				</div>
			</CardContent>
		</Card>
	);
}
