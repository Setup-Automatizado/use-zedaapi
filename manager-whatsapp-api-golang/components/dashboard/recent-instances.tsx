/**
 * Recent Instances Component
 *
 * Displays a minimal list of recent WhatsApp instances with avatars,
 * connection status, phone numbers, and metadata in a clean row layout.
 */

import { ArrowRight, Inbox } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardAction,
	CardContent,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { INSTANCE_STATUS, type InstanceStatus } from "@/lib/constants";
import { cn } from "@/lib/utils";
import type { DeviceInfo, Instance } from "@/types";

export interface DeviceMap {
	[instanceId: string]: DeviceInfo | null;
}

export interface RecentInstancesProps {
	instances: Instance[];
	deviceMap?: DeviceMap;
}

function getInstanceStatus(instance: Instance): InstanceStatus {
	if (instance.whatsappConnected && instance.phoneConnected) {
		return INSTANCE_STATUS.CONNECTED;
	}
	if (!instance.whatsappConnected || !instance.phoneConnected) {
		return INSTANCE_STATUS.DISCONNECTED;
	}
	return INSTANCE_STATUS.PENDING;
}

function formatRelativeDate(dateString: string | undefined): string {
	if (!dateString) return "-";
	const date = new Date(dateString);
	if (isNaN(date.getTime())) return "-";
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
	const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
	const diffMinutes = Math.floor(diffMs / (1000 * 60));

	if (diffMinutes < 1) return "now";
	if (diffMinutes < 60) return `${diffMinutes}min`;
	if (diffHours < 24) return `${diffHours}h`;
	if (diffDays < 7) return `${diffDays}d`;
	if (diffDays < 30) {
		const weeks = Math.floor(diffDays / 7);
		return `${weeks}w`;
	}
	const months = Math.floor(diffDays / 30);
	return `${months}mo`;
}

function formatPhoneNumber(phone: string | undefined): string {
	if (!phone) return "-";
	const cleaned = phone.replace(/\D/g, "");

	// Brazilian: +55 DD NNNNN-NNNN
	if (cleaned.startsWith("55") && cleaned.length >= 12) {
		const areaCode = cleaned.substring(2, 4);
		const localNumber = cleaned.substring(4);
		if (localNumber.length === 9) {
			const firstPart = localNumber.substring(0, 5);
			const secondPart = localNumber.substring(5);
			return `+55 ${areaCode} ${firstPart}-${secondPart}`;
		}
		if (localNumber.length === 8) {
			const firstPart = localNumber.substring(0, 4);
			const secondPart = localNumber.substring(4);
			return `+55 ${areaCode} ${firstPart}-${secondPart}`;
		}
	}

	// Argentina mobile: +54 9 XXXX XX-XXXX
	if (cleaned.startsWith("549") && cleaned.length >= 13) {
		const areaCode = cleaned.substring(3, 7);
		const localNumber = cleaned.substring(7);
		if (localNumber.length >= 6) {
			const firstPart = localNumber.substring(0, 2);
			const secondPart = localNumber.substring(2);
			return `+54 9 ${areaCode} ${firstPart}-${secondPart}`;
		}
	}

	// USA/Canada: +1 XXX XXX-XXXX
	if (cleaned.startsWith("1") && cleaned.length === 11) {
		const areaCode = cleaned.substring(1, 4);
		const firstPart = cleaned.substring(4, 7);
		const secondPart = cleaned.substring(7);
		return `+1 ${areaCode} ${firstPart}-${secondPart}`;
	}

	// Default international format
	if (cleaned.length > 6) {
		const countryCode = cleaned.substring(0, 2);
		const rest = cleaned.substring(2);
		return `+${countryCode} ${rest.replace(/(\d{4,5})(\d{4})$/, "$1-$2")}`;
	}

	return phone;
}

function getInitials(name: string): string {
	return name
		.split(" ")
		.map((word) => word[0])
		.join("")
		.toUpperCase()
		.slice(0, 2);
}

function getAvatarColor(name: string): string {
	const colors = [
		"bg-emerald-500",
		"bg-blue-500",
		"bg-violet-500",
		"bg-amber-500",
		"bg-rose-500",
		"bg-cyan-500",
		"bg-indigo-500",
		"bg-pink-500",
	];
	const index =
		name.split("").reduce((acc, char) => acc + char.charCodeAt(0), 0) %
		colors.length;
	return colors[index];
}

export function RecentInstances({
	instances,
	deviceMap = {},
}: RecentInstancesProps) {
	const recentInstances = instances.slice(0, 10);

	return (
		<Card>
			<CardHeader className="pb-3">
				<CardTitle className="text-base font-medium">
					Recent Instances
				</CardTitle>
				<CardAction>
					<Button variant="ghost" size="sm" asChild>
						<Link href="/instances">
							View all
							<ArrowRight className="h-4 w-4" data-icon="inline-end" />
						</Link>
					</Button>
				</CardAction>
			</CardHeader>
			<CardContent className="pt-0">
				{recentInstances.length === 0 ? (
					<div className="flex flex-col items-center justify-center py-12 text-center">
						<div className="mb-3 flex h-12 w-12 items-center justify-center rounded-2xl bg-muted">
							<Inbox className="h-6 w-6 text-muted-foreground" />
						</div>
						<p className="text-sm font-medium text-foreground">
							No instances registered
						</p>
						<p className="mt-1 text-sm text-muted-foreground">
							Create your first instance to get started
						</p>
						<Button size="sm" className="mt-4" asChild>
							<Link href="/instances/new">Create Instance</Link>
						</Button>
					</div>
				) : (
					<div className="divide-y divide-border/40">
						{recentInstances.map((instance) => {
							const status = getInstanceStatus(instance);
							const isConnected = status === INSTANCE_STATUS.CONNECTED;
							const deviceInfo = deviceMap[instance.instanceId];

							// Get phone from device info or storeJid
							const phone =
								deviceInfo?.phone || instance.storeJid?.split("@")[0];

							// Get avatar URL from device info
							const avatarUrl = deviceInfo?.imgUrl;

							// Get display name (prefer device name if available)
							const displayName = deviceInfo?.name || instance.name;

							return (
								<Link
									key={instance.id}
									href={`/instances/${instance.id}`}
									className="group flex items-center gap-3 px-1 py-3 transition-colors hover:bg-muted/40 -mx-1 first:pt-0 last:pb-0"
								>
									{/* Avatar */}
									<div className="relative flex-shrink-0">
										<Avatar className="h-9 w-9">
											{avatarUrl ? (
												<AvatarImage
													src={avatarUrl}
													alt={displayName}
													className="object-cover"
												/>
											) : null}
											<AvatarFallback
												className={cn(
													"text-xs font-medium text-white",
													getAvatarColor(instance.name),
												)}
											>
												{getInitials(instance.name)}
											</AvatarFallback>
										</Avatar>
										{isConnected && (
											<span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-card bg-emerald-500" />
										)}
									</div>

									{/* Name */}
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-foreground group-hover:text-primary">
											{instance.name}
										</p>
									</div>

									{/* Phone - hidden on mobile */}
									<div className="hidden sm:block min-w-[140px] text-right">
										<span className="text-xs font-mono text-muted-foreground">
											{formatPhoneNumber(phone)}
										</span>
									</div>

									{/* Status */}
									<div className="flex items-center gap-1.5 min-w-[90px]">
										<span
											className={cn(
												"h-1.5 w-1.5 rounded-full",
												isConnected ? "bg-emerald-500" : "bg-red-400",
											)}
										/>
										<span
											className={cn(
												"text-xs",
												isConnected
													? "text-emerald-600 dark:text-emerald-400"
													: "text-red-500 dark:text-red-400",
											)}
										>
											{isConnected ? "Connected" : "Disconnected"}
										</span>
									</div>

									{/* Middleware - hidden on tablet and mobile */}
									<span className="hidden lg:block text-xs text-muted-foreground w-10 text-center capitalize">
										{instance.middleware || "web"}
									</span>

									{/* Time */}
									<span className="text-xs text-muted-foreground w-10 text-right">
										{formatRelativeDate(instance.created)}
									</span>
								</Link>
							);
						})}
					</div>
				)}
			</CardContent>
		</Card>
	);
}
