/**
 * Instance Filter Component
 *
 * Combobox for filtering metrics by instance ID with friendly names,
 * phone numbers, and avatar images.
 *
 * @module components/metrics/instance-filter
 */

"use client";

import { Check, ChevronsUpDown, Phone, Server, User } from "lucide-react";
import * as React from "react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
	Command,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "@/components/ui/command";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { useInstanceNames } from "@/hooks/use-instance-names";
import { formatPhoneNumber } from "@/lib/phone";
import { cn } from "@/lib/utils";

export interface InstanceFilterProps {
	/** Available instance IDs */
	instances: string[];
	/** Currently selected instance (null for all) */
	selectedInstance: string | null;
	/** Callback when selection changes */
	onSelect: (instanceId: string | null) => void;
	/** Additional CSS classes */
	className?: string;
	/** Disabled state */
	disabled?: boolean;
}

export function InstanceFilter({
	instances,
	selectedInstance,
	onSelect,
	className,
	disabled = false,
}: InstanceFilterProps) {
	const [open, setOpen] = React.useState(false);
	const { getDisplayName, getInstanceInfo } = useInstanceNames();

	const selectedInfo = selectedInstance
		? getInstanceInfo(selectedInstance)
		: null;

	const displayValue = selectedInfo
		? selectedInfo.name
		: selectedInstance
			? getDisplayName(selectedInstance)
			: "All instances";

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<Button
					variant="outline"
					aria-expanded={open}
					className={cn("w-[280px] justify-between h-10", className)}
					disabled={disabled || instances.length === 0}
				>
					<div className="flex items-center gap-2 truncate">
						{selectedInfo?.avatarUrl ? (
							<Avatar className="h-6 w-6">
								<AvatarImage src={selectedInfo.avatarUrl} alt={selectedInfo.name} />
								<AvatarFallback className="text-xs">
									{selectedInfo.name.slice(0, 2).toUpperCase()}
								</AvatarFallback>
							</Avatar>
						) : (
							<Server className="h-4 w-4 shrink-0 opacity-50" />
						)}
						<div className="flex flex-col items-start min-w-0">
							<span className="truncate text-sm">{displayValue}</span>
							{selectedInfo?.formattedPhone && (
								<span className="truncate text-xs text-muted-foreground">
									{selectedInfo.formattedPhone}
								</span>
							)}
						</div>
					</div>
					<ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
				</Button>
			</PopoverTrigger>
			<PopoverContent className="w-[320px] p-0" align="start">
				<Command>
					<CommandInput placeholder="Search by name or phone..." />
					<CommandList>
						<CommandEmpty>No instance found.</CommandEmpty>
						<CommandGroup>
							{/* All instances option */}
							<CommandItem
								value="all"
								onSelect={() => {
									onSelect(null);
									setOpen(false);
								}}
							>
								<Check
									className={cn(
										"mr-2 h-4 w-4",
										selectedInstance === null ? "opacity-100" : "opacity-0",
									)}
								/>
								<Server className="mr-2 h-4 w-4 opacity-50" />
								All instances
							</CommandItem>

							{/* Individual instances */}
							{instances.map((instanceId) => {
								const info = getInstanceInfo(instanceId);
								const displayName = info?.name || getDisplayName(instanceId);
								const formattedPhone = info?.formattedPhone;
								const avatarUrl = info?.avatarUrl;
								const isSelected = selectedInstance === instanceId;

								return (
									<CommandItem
										key={instanceId}
										value={`${instanceId} ${displayName} ${info?.phone || ""}`}
										onSelect={() => {
											onSelect(instanceId);
											setOpen(false);
										}}
										className="py-2"
									>
										<Check
											className={cn(
												"mr-2 h-4 w-4 shrink-0",
												isSelected ? "opacity-100" : "opacity-0",
											)}
										/>
										<Avatar className="h-8 w-8 mr-2">
											{avatarUrl ? (
												<AvatarImage src={avatarUrl} alt={displayName} />
											) : null}
											<AvatarFallback className="text-xs bg-muted">
												{displayName.slice(0, 2).toUpperCase()}
											</AvatarFallback>
										</Avatar>
										<div className="flex flex-col min-w-0 flex-1">
											<span className="truncate font-medium">
												{displayName}
											</span>
											{formattedPhone ? (
												<span className="truncate text-xs text-muted-foreground flex items-center gap-1">
													<Phone className="h-3 w-3" />
													{formattedPhone}
												</span>
											) : (
												<span className="truncate text-xs text-muted-foreground font-mono">
													{truncateInstanceId(instanceId)}
												</span>
											)}
										</div>
										{info?.connected && (
											<span className="ml-2 h-2 w-2 rounded-full bg-emerald-500" title="Connected" />
										)}
									</CommandItem>
								);
							})}
						</CommandGroup>
					</CommandList>
				</Command>
			</PopoverContent>
		</Popover>
	);
}

/**
 * Truncate instance ID for display
 */
function truncateInstanceId(id: string): string {
	if (id.length <= 12) return id;
	return `${id.slice(0, 8)}...${id.slice(-4)}`;
}

/**
 * Instance Badge
 *
 * Displays a compact instance identifier with avatar, name, and phone
 */
export function InstanceBadge({
	instanceId,
	className,
	showPhone = false,
	size = "sm",
}: {
	instanceId: string;
	className?: string;
	showPhone?: boolean;
	size?: "sm" | "md" | "lg";
}) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();
	const info = getInstanceInfo(instanceId);
	const displayName = info?.name || getDisplayName(instanceId);

	const sizeClasses = {
		sm: "px-2 py-1 text-xs",
		md: "px-2.5 py-1.5 text-sm",
		lg: "px-3 py-2 text-sm",
	};

	const avatarSizes = {
		sm: "h-4 w-4",
		md: "h-5 w-5",
		lg: "h-6 w-6",
	};

	return (
		<span
			className={cn(
				"inline-flex items-center gap-1.5 rounded-md bg-muted",
				sizeClasses[size],
				className,
			)}
			title={instanceId}
		>
			<Avatar className={avatarSizes[size]}>
				{info?.avatarUrl ? (
					<AvatarImage src={info.avatarUrl} alt={displayName} />
				) : null}
				<AvatarFallback className="text-[10px] bg-background">
					{displayName.slice(0, 2).toUpperCase()}
				</AvatarFallback>
			</Avatar>
			<span className="font-medium">{displayName}</span>
			{showPhone && info?.formattedPhone && (
				<span className="text-muted-foreground">
					({info.formattedPhone})
				</span>
			)}
		</span>
	);
}

/**
 * Instance Card Display
 *
 * Shows instance with avatar, name, phone in a card-like format
 */
export function InstanceDisplay({
	instanceId,
	className,
	showId = false,
}: {
	instanceId: string;
	className?: string;
	showId?: boolean;
}) {
	const { getDisplayName, getInstanceInfo } = useInstanceNames();
	const info = getInstanceInfo(instanceId);
	const displayName = info?.name || getDisplayName(instanceId);

	return (
		<div className={cn("flex items-center gap-3", className)}>
			<Avatar className="h-10 w-10">
				{info?.avatarUrl ? (
					<AvatarImage src={info.avatarUrl} alt={displayName} />
				) : null}
				<AvatarFallback className="text-sm bg-muted">
					<User className="h-5 w-5 opacity-50" />
				</AvatarFallback>
			</Avatar>
			<div className="flex flex-col min-w-0">
				<span className="font-medium truncate">{displayName}</span>
				{info?.formattedPhone ? (
					<span className="text-sm text-muted-foreground flex items-center gap-1">
						<Phone className="h-3 w-3" />
						{info.formattedPhone}
					</span>
				) : showId ? (
					<span className="text-xs text-muted-foreground font-mono">
						{truncateInstanceId(instanceId)}
					</span>
				) : null}
			</div>
			{info?.connected && (
				<span className="ml-auto h-2.5 w-2.5 rounded-full bg-emerald-500" title="Connected" />
			)}
		</div>
	);
}
