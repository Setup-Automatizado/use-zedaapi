/**
 * Recent Instances Component
 *
 * Displays a minimal list of recent WhatsApp instances with avatars,
 * connection status, phone numbers, and metadata in a clean row layout.
 */

import { ArrowRight, Inbox } from "lucide-react";
import Link from "next/link";
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
import { formatPhoneNumber } from "@/lib/phone";
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
	if (Number.isNaN(date.getTime())) return "-";
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
	const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
	const diffMinutes = Math.floor(diffMs / (1000 * 60));

	if (diffMinutes < 1) return "now";
	if (diffMinutes < 60) return `${diffMinutes}m`;
	if (diffHours < 24) return `${diffHours}h`;
	if (diffDays < 7) return `${diffDays}d`;
	if (diffDays < 30) {
		const weeks = Math.floor(diffDays / 7);
		return `${weeks}w`;
	}
	const months = Math.floor(diffDays / 30);
	return `${months}mo`;
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
			<CardHeader>
				<CardTitle className="text-base font-medium">
					Recent Instances
				</CardTitle>
				<CardAction>
					<Button variant="ghost" size="sm" asChild>
						<Link href="/instances">
							View all
							<ArrowRight className="h-4 w-4" />
						</Link>
					</Button>
				</CardAction>
			</CardHeader>
			<CardContent>
				{recentInstances.length === 0 ? (
					<div className="flex flex-col items-center justify-center py-12 text-center">
						<div className="mb-4 flex h-14 w-14 items-center justify-center rounded-full bg-muted">
							<Inbox className="h-7 w-7 text-muted-foreground" />
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
					<div className="space-y-1">
						{recentInstances.map((instance) => {
							const status = getInstanceStatus(instance);
							const isConnected = status === INSTANCE_STATUS.CONNECTED;
							const deviceInfo = deviceMap[instance.instanceId];
							const phone =
								deviceInfo?.phone || instance.storeJid?.split("@")[0];
							const avatarUrl = deviceInfo?.imgUrl;

							return (
								<Link
									key={instance.id}
									href={`/instances/${instance.id}`}
									className="group flex items-center gap-4 rounded-lg px-3 py-2.5 transition-colors hover:bg-muted/50"
								>
									{/* Avatar with status indicator */}
									<div className="relative shrink-0">
										<Avatar className="h-10 w-10 ring-2 ring-background">
											{avatarUrl ? (
												<AvatarImage
													src={avatarUrl}
													alt={instance.name}
													className="object-cover"
												/>
											) : null}
											<AvatarFallback
												className={cn(
													"text-xs font-semibold text-white",
													getAvatarColor(instance.name),
												)}
											>
												{getInitials(instance.name)}
											</AvatarFallback>
										</Avatar>
										<span
											className={cn(
												"absolute bottom-0 right-0 h-3 w-3 rounded-full ring-2 ring-background",
												isConnected ? "bg-emerald-500" : "bg-zinc-400",
											)}
										/>
									</div>

									{/* Name and phone */}
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-foreground group-hover:text-primary transition-colors">
											{instance.name}
										</p>
										{phone && (
											<p className="truncate text-xs text-muted-foreground font-mono">
												{formatPhoneNumber(phone)}
											</p>
										)}
									</div>

									{/* Status badge */}
									<div className="hidden sm:flex items-center gap-1.5 shrink-0">
										<span
											className={cn(
												"inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium",
												isConnected
													? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
													: "bg-zinc-500/10 text-zinc-600 dark:text-zinc-400",
											)}
										>
											<span
												className={cn(
													"h-1.5 w-1.5 rounded-full",
													isConnected ? "bg-emerald-500" : "bg-zinc-400",
												)}
											/>
											{isConnected ? "Connected" : "Disconnected"}
										</span>
									</div>

									{/* Time */}
									<span className="text-xs text-muted-foreground shrink-0 tabular-nums">
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
