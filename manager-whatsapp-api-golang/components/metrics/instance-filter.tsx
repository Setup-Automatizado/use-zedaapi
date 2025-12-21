/**
 * Instance Filter Component
 *
 * Combobox for filtering metrics by instance ID.
 *
 * @module components/metrics/instance-filter
 */

"use client";

import { Check, ChevronsUpDown, Server } from "lucide-react";
import * as React from "react";
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

	const displayValue = selectedInstance
		? truncateInstanceId(selectedInstance)
		: "All instances";

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<Button
					variant="outline"
					aria-expanded={open}
					className={cn("w-[200px] justify-between h-9", className)}
					disabled={disabled || instances.length === 0}
				>
					<div className="flex items-center gap-2 truncate">
						<Server className="h-4 w-4 shrink-0 opacity-50" />
						<span className="truncate">{displayValue}</span>
					</div>
					<ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
				</Button>
			</PopoverTrigger>
			<PopoverContent className="w-[200px] p-0" align="start">
				<Command>
					<CommandInput placeholder="Search instance..." />
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
								All instances
							</CommandItem>

							{/* Individual instances */}
							{instances.map((instanceId) => (
								<CommandItem
									key={instanceId}
									value={instanceId}
									onSelect={() => {
										onSelect(instanceId);
										setOpen(false);
									}}
								>
									<Check
										className={cn(
											"mr-2 h-4 w-4",
											selectedInstance === instanceId
												? "opacity-100"
												: "opacity-0",
										)}
									/>
									<span className="truncate font-mono text-xs">
										{truncateInstanceId(instanceId)}
									</span>
								</CommandItem>
							))}
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
 * Displays a compact instance identifier
 */
export function InstanceBadge({
	instanceId,
	className,
}: {
	instanceId: string;
	className?: string;
}) {
	return (
		<span
			className={cn(
				"inline-flex items-center gap-1 rounded-md bg-muted px-2 py-1 text-xs font-mono",
				className,
			)}
			title={instanceId}
		>
			<Server className="h-3 w-3 opacity-50" />
			{truncateInstanceId(instanceId)}
		</span>
	);
}
