"use client";

import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import { Calendar, Phone } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { formatPhoneNumber } from "@/lib/phone";
import type { DeviceInfo, Instance } from "@/types";
import { InstanceActionsDropdown } from "./instance-actions-dropdown";
import { InstanceStatusBadge } from "./instance-status-badge";

export interface InstanceCardProps {
	instance: Instance;
	deviceInfo?: DeviceInfo | null;
	onRestart?: () => void | Promise<void>;
	onDisconnect?: () => void | Promise<void>;
	onDelete?: () => void | Promise<void>;
}

export function InstanceCard({
	instance,
	deviceInfo,
	onRestart,
	onDisconnect,
	onDelete,
}: InstanceCardProps) {
	// Get phone from device info (preferred) or storeJid as fallback
	const phone = deviceInfo?.phone || instance.storeJid?.split("@")[0];
	const hasActiveSubscription = instance.subscriptionActive;
	const dueDate = React.useMemo(
		() => (instance.due ? new Date(instance.due) : null),
		[instance.due],
	);
	const [now] = React.useState(() => Date.now());
	const isExpiringSoon = React.useMemo(() => {
		if (!hasActiveSubscription || !dueDate) return false;
		return dueDate.getTime() - now < 7 * 24 * 60 * 60 * 1000;
	}, [hasActiveSubscription, dueDate, now]);

	const createdDate = React.useMemo(() => {
		try {
			if (!instance.created) {
				return "-";
			}
			const date = new Date(instance.created);
			if (isNaN(date.getTime())) {
				return "-";
			}
			return formatDistanceToNow(date, {
				addSuffix: true,
				locale: enUS,
			});
		} catch {
			return "-";
		}
	}, [instance.created]);

	return (
		<Card>
			<CardHeader className="flex flex-row items-start justify-between space-y-0 pb-3">
				<div className="space-y-1 flex-1">
					<Link
						href={`/instances/${instance.id}`}
						className="font-semibold hover:underline"
					>
						{instance.name}
					</Link>
					<p className="text-xs text-muted-foreground">
						{instance.sessionName}
					</p>
				</div>
				<InstanceActionsDropdown
					instance={instance}
					onRestart={onRestart}
					onDisconnect={onDisconnect}
					onDelete={onDelete}
				/>
			</CardHeader>
			<CardContent className="space-y-3">
				<div className="flex items-center justify-between">
					<span className="text-sm text-muted-foreground">Status</span>
					<InstanceStatusBadge
						connected={instance.whatsappConnected}
						smartphoneConnected={instance.phoneConnected}
					/>
				</div>

				{phone && (
					<div className="flex items-center justify-between">
						<span className="text-sm text-muted-foreground flex items-center gap-1.5">
							<Phone className="h-3.5 w-3.5" />
							Phone
						</span>
						<span className="font-mono text-sm">
							{formatPhoneNumber(phone, "-")}
						</span>
					</div>
				)}

				<div className="flex items-center justify-between">
					<span className="text-sm text-muted-foreground">Subscription</span>
					{hasActiveSubscription ? (
						<Badge variant={isExpiringSoon ? "outline" : "secondary"}>
							Active
						</Badge>
					) : (
						<Badge variant="destructive">Inactive</Badge>
					)}
				</div>

				<div className="flex items-center justify-between pt-2 border-t">
					<span className="text-xs text-muted-foreground flex items-center gap-1.5">
						<Calendar className="h-3 w-3" />
						Created
					</span>
					<span className="text-xs text-muted-foreground">{createdDate}</span>
				</div>
			</CardContent>
		</Card>
	);
}
