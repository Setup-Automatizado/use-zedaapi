"use client";

import {
	Eye,
	LogOut,
	MoreVertical,
	RotateCcw,
	Settings,
	Trash,
	Webhook,
} from "lucide-react";

import Link from "next/link";
import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { Instance } from "@/types";

export interface InstanceActionsDropdownProps {
	instance: Instance;
	onRestart?: () => void | Promise<void>;
	onDisconnect?: () => void | Promise<void>;
	onDelete?: () => void | Promise<void>;
}

export function InstanceActionsDropdown({
	instance,
	onRestart,
	onDisconnect,
	onDelete,
}: InstanceActionsDropdownProps) {
	const [showDeleteDialog, setShowDeleteDialog] = useState(false);
	const [isOpen, setIsOpen] = useState(false);

	const handleDelete = async () => {
		if (onDelete) {
			await onDelete();
			setShowDeleteDialog(false);
			setIsOpen(false);
		}
	};

	const handleRestart = async () => {
		if (onRestart) {
			await onRestart();
			setIsOpen(false);
		}
	};

	const handleDisconnect = async () => {
		if (onDisconnect) {
			await onDisconnect();
			setIsOpen(false);
		}
	};

	return (
		<>
			<DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
				<DropdownMenuTrigger asChild>
					<Button variant="ghost" size="icon-sm">
						<MoreVertical className="h-4 w-4" />
						<span className="sr-only">Open menu</span>
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent align="end">
					<DropdownMenuItem asChild>
						<Link href={`/instances/${instance.id}`}>
							<Eye className="h-4 w-4" />
							View Details
						</Link>
					</DropdownMenuItem>
					<DropdownMenuItem asChild>
						<Link href={`/instances/${instance.id}?tab=webhooks`}>
							<Webhook className="h-4 w-4" />
							Webhooks
						</Link>
					</DropdownMenuItem>
					<DropdownMenuItem asChild>
						<Link href={`/instances/${instance.id}?tab=settings`}>
							<Settings className="h-4 w-4" />
							Settings
						</Link>
					</DropdownMenuItem>
					{onRestart && (
						<DropdownMenuItem onClick={handleRestart}>
							<RotateCcw className="h-4 w-4" />
							Restart
						</DropdownMenuItem>
					)}
					{onDisconnect && (
						<DropdownMenuItem onClick={handleDisconnect}>
							<LogOut className="h-4 w-4" />
							Disconnect
						</DropdownMenuItem>
					)}
					{onDelete && (
						<>
							<DropdownMenuSeparator />
							<DropdownMenuItem
								variant="destructive"
								onClick={() => setShowDeleteDialog(true)}
							>
								<Trash className="h-4 w-4" />
								Delete
							</DropdownMenuItem>
						</>
					)}
				</DropdownMenuContent>
			</DropdownMenu>

			{onDelete && (
				<ConfirmDialog
					open={showDeleteDialog}
					onOpenChange={setShowDeleteDialog}
					title="Delete instance"
					description={`Are you sure you want to delete the instance "${instance.name}"? This action cannot be undone.`}
					confirmLabel="Delete"
					cancelLabel="Cancel"
					variant="destructive"
					onConfirm={handleDelete}
				/>
			)}
		</>
	);
}
