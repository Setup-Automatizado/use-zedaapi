"use client";

import * as React from "react";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Loader2 } from "lucide-react";

export interface ConfirmDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	description: string;
	confirmLabel?: string;
	cancelLabel?: string;
	variant?: "default" | "destructive";
	onConfirm: () => void | Promise<void>;
	isLoading?: boolean;
}

export function ConfirmDialog({
	open,
	onOpenChange,
	title,
	description,
	confirmLabel = "Confirm",
	cancelLabel = "Cancel",
	variant = "default",
	onConfirm,
	isLoading = false,
}: ConfirmDialogProps) {
	const [isConfirming, setIsConfirming] = React.useState(false);

	const handleConfirm = async () => {
		try {
			setIsConfirming(true);
			await onConfirm();
			onOpenChange(false);
		} catch (error) {
			console.error("Confirm action failed:", error);
		} finally {
			setIsConfirming(false);
		}
	};

	const loading = isLoading || isConfirming;

	return (
		<AlertDialog open={open} onOpenChange={onOpenChange}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>{title}</AlertDialogTitle>
					<AlertDialogDescription>
						{description}
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel disabled={loading}>
						{cancelLabel}
					</AlertDialogCancel>
					<AlertDialogAction
						variant={variant}
						onClick={(e) => {
							e.preventDefault();
							handleConfirm();
						}}
						disabled={loading}
					>
						{loading && (
							<Loader2 className="mr-2 h-4 w-4 animate-spin" />
						)}
						{confirmLabel}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
