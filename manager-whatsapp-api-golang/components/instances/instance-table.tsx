"use client";

import * as React from "react";
import { Instance, DeviceInfo } from "@/types";
import { DataTable, type Column } from "@/components/shared/data-table";
import { InstanceStatusBadge } from "./instance-status-badge";
import { InstanceActionsDropdown } from "./instance-actions-dropdown";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { formatDistanceToNow } from "date-fns";
import { enUS } from "date-fns/locale";
import { cn } from "@/lib/utils";

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

export function InstanceTable({
	instances,
	deviceMap = {},
	isLoading = false,
	pagination,
	onRestart,
	onDisconnect,
	onDelete,
}: InstanceTableProps) {
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
						<div className="relative flex-shrink-0">
							<Avatar className="h-8 w-8">
								{avatarUrl ? (
									<AvatarImage
										src={avatarUrl}
										alt={instance.name}
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
								<span className="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-background bg-emerald-500" />
							)}
						</div>
						<div className="flex flex-col min-w-0">
							<span className="font-medium truncate">
								{instance.name}
							</span>
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
				// Get phone from device info (preferred) or storeJid as fallback
				const deviceInfo = deviceMap[instance.instanceId];
				const phone =
					deviceInfo?.phone || instance.storeJid?.split("@")[0];

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
					dueDate &&
					dueDate.getTime() - Date.now() < 7 * 24 * 60 * 60 * 1000;

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
					if (isNaN(date.getTime())) {
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
			render: (instance) => (
				<InstanceActionsDropdown
					instance={instance}
					onRestart={
						onRestart ? () => onRestart(instance) : undefined
					}
					onDisconnect={
						onDisconnect ? () => onDisconnect(instance) : undefined
					}
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
		/>
	);
}
