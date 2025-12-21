"use client";

import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import { useRouter } from "next/navigation";
import { type Column, DataTable } from "@/components/shared/data-table";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { formatPhoneNumber } from "@/lib/phone";
import { cn } from "@/lib/utils";
import type { DeviceInfo, Instance } from "@/types";
import { InstanceActionsDropdown } from "./instance-actions-dropdown";
import { InstanceStatusBadge } from "./instance-status-badge";

export interface DeviceMap {
	[instanceId: string]: DeviceInfo | null;
}

export interface InstanceTableProps {
	instances: Instance[];
	deviceMap?: DeviceMap;
	isLoading?: boolean;
	pagination?: {
		page: number;
		pageSize: number;
		total: number;
		onPageChange: (page: number) => void;
	};
	onRestart?: (instance: Instance) => void | Promise<void>;
	onDisconnect?: (instance: Instance) => void | Promise<void>;
	onDelete?: (instance: Instance) => void | Promise<void>;
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

export function InstanceTable({
	instances,
	deviceMap = {},
	isLoading = false,
	pagination,
	onRestart,
	onDisconnect,
	onDelete,
}: InstanceTableProps) {
	const router = useRouter();

	const handleRowClick = (instance: Instance) => {
		router.push(`/instances/${instance.id}`);
	};

	const columns: Column<Instance>[] = [
		{
			key: "name",
			label: "Name",
			render: (instance) => {
				const deviceInfo = deviceMap[instance.instanceId];
				const avatarUrl = deviceInfo?.imgUrl;
				const isConnected =
					instance.whatsappConnected && instance.phoneConnected;

				return (
					<div className="flex items-center gap-3">
						<div className="relative shrink-0">
							<Avatar className="h-9 w-9 ring-2 ring-background">
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
									"absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full ring-2 ring-background",
									isConnected ? "bg-emerald-500" : "bg-zinc-400",
								)}
							/>
						</div>
						<div className="flex flex-col min-w-0">
							<span className="font-medium truncate">{instance.name}</span>
							<span className="text-xs text-muted-foreground truncate">
								{instance.sessionName}
							</span>
						</div>
					</div>
				);
			},
		},
		{
			key: "status",
			label: "Status",
			render: (instance) => (
				<InstanceStatusBadge
					connected={instance.whatsappConnected}
					smartphoneConnected={instance.phoneConnected}
				/>
			),
		},
		{
			key: "phone",
			label: "Phone",
			render: (instance) => {
				const deviceInfo = deviceMap[instance.instanceId];
				const phone = deviceInfo?.phone || instance.storeJid?.split("@")[0];

				if (phone) {
					return (
						<span className="font-mono text-sm">
							{formatPhoneNumber(phone)}
						</span>
					);
				}
				return <span className="text-muted-foreground">-</span>;
			},
		},
		{
			key: "subscription",
			label: "Subscription",
			render: (instance) => {
				if (!instance.subscriptionActive) {
					return <Badge variant="destructive">Inactive</Badge>;
				}

				const dueDate = instance.due ? new Date(instance.due) : null;
				const isExpiringSoon =
					dueDate && dueDate.getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000;

				return (
					<Badge variant={isExpiringSoon ? "outline" : "secondary"}>
						Active
					</Badge>
				);
			},
		},
		{
			key: "created",
			label: "Created",
			render: (instance) => {
				try {
					if (!instance.created) {
						return <span className="text-muted-foreground">-</span>;
					}
					const date = new Date(instance.created);
					if (Number.isNaN(date.getTime())) {
						return <span className="text-muted-foreground">-</span>;
					}
					return (
						<span className="text-sm text-muted-foreground">
							{formatDistanceToNow(date, {
								addSuffix: true,
								locale: enUS,
							})}
						</span>
					);
				} catch {
					return <span className="text-muted-foreground">-</span>;
				}
			},
		},
		{
			key: "actions",
			label: "",
			className: "w-[50px]",
			preventRowClick: true,
			render: (instance) => (
				<InstanceActionsDropdown
					instance={instance}
					onRestart={onRestart ? () => onRestart(instance) : undefined}
					onDisconnect={onDisconnect ? () => onDisconnect(instance) : undefined}
					onDelete={onDelete ? () => onDelete(instance) : undefined}
				/>
			),
		},
	];

	return (
		<DataTable
			columns={columns}
			data={instances}
			isLoading={isLoading}
			emptyMessage="No instances found"
			emptyDescription="Create your first instance to get started"
			pagination={pagination}
			getRowKey={(instance) => instance.id}
			onRowClick={handleRowClick}
		/>
	);
}
