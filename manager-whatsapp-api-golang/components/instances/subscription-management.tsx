/**
 * Subscription Management Component
 *
 * Manages instance subscription status with activate/cancel actions.
 * Displays subscription status, partner info, and action buttons.
 *
 * @example
 * ```tsx
 * <SubscriptionManagement instance={instance} onUpdate={mutate} />
 * ```
 */

"use client";

import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { activateSubscription, cancelSubscription } from "@/actions";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import type { Instance } from "@/types";
import { isError } from "@/types";

export interface SubscriptionManagementProps {
	/** Instance to manage subscription for */
	instance: Instance;

	/** Callback after subscription changes */
	onUpdate: () => void;
}

export function SubscriptionManagement({
	instance,
	onUpdate,
}: SubscriptionManagementProps) {
	const [isActivating, setIsActivating] = useState(false);
	const [isCanceling, setIsCanceling] = useState(false);

	const isCanceled = Boolean(instance.canceledAt);
	const isActive = instance.subscriptionActive && !isCanceled;
	const isPending = !instance.subscriptionActive && !isCanceled;

	const handleActivate = async () => {
		setIsActivating(true);
		const result = await activateSubscription(instance.id, instance.token);

		if (isError(result)) {
			toast.error("Activation failed", {
				description: result.error || "Unknown error",
			});
			setIsActivating(false);
			return;
		}

		toast.success("Subscription activated!", {
			description: "The subscription has been activated successfully.",
		});
		onUpdate();
		setIsActivating(false);
	};

	const handleCancel = async () => {
		setIsCanceling(true);
		const result = await cancelSubscription(instance.id, instance.token);

		if (isError(result)) {
			toast.error("Cancellation failed", {
				description: result.error || "Unknown error",
			});
			setIsCanceling(false);
			return;
		}

		toast.success("Subscription canceled", {
			description: "The subscription has been canceled successfully.",
		});
		onUpdate();
		setIsCanceling(false);
	};

	return (
		<Card>
			<CardHeader>
				<div className="flex items-center justify-between">
					<div>
						<CardTitle>Subscription</CardTitle>
						<CardDescription>
							Manage the subscription for this instance
						</CardDescription>
					</div>
					{isActive && (
						<Badge
							variant="outline"
							className="bg-green-50 text-green-700 border-green-200"
						>
							<CheckCircle2 className="mr-1 h-3 w-3" />
							Active
						</Badge>
					)}
					{isPending && (
						<Badge
							variant="outline"
							className="bg-yellow-50 text-yellow-700 border-yellow-200"
						>
							Pending
						</Badge>
					)}
					{isCanceled && (
						<Badge
							variant="outline"
							className="bg-gray-50 text-gray-700 border-gray-200"
						>
							<XCircle className="mr-1 h-3 w-3" />
							Canceled
						</Badge>
					)}
				</div>
			</CardHeader>
			<CardContent>
				<div className="space-y-4">
					<div>
						<p className="text-sm text-muted-foreground mb-2">
							Status:{" "}
							<span className="font-semibold">
								{isActive ? "Active" : isCanceled ? "Canceled" : "Pending"}
							</span>
						</p>
						{instance.due && (
							<p className="text-sm text-muted-foreground mb-2">
								Due Date:{" "}
								<span className="font-mono">
									{new Date(instance.due).toLocaleDateString()}
								</span>
							</p>
						)}
					</div>

					<div className="flex gap-2">
						{!isActive && (
							<Button
								onClick={handleActivate}
								disabled={isActivating}
								className="flex-1"
							>
								{isActivating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
								Activate Subscription
							</Button>
						)}

						{isActive && (
							<Button
								variant="destructive"
								onClick={handleCancel}
								disabled={isCanceling}
								className="flex-1"
							>
								{isCanceling && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
								Cancel Subscription
							</Button>
						)}
					</div>
				</div>
			</CardContent>
		</Card>
	);
}
