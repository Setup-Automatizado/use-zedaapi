"use client";

import type { LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AnimateIn } from "@/components/shared/motion";
import { cn } from "@/lib/utils";

interface EmptyStateProps {
	icon: LucideIcon;
	title: string;
	description: string;
	actionLabel?: string;
	onAction?: () => void;
	secondaryAction?: { label: string; onClick: () => void };
	variant?: "zero" | "no-results";
}

export function EmptyState({
	icon: Icon,
	title,
	description,
	actionLabel,
	onAction,
	secondaryAction,
	variant = "zero",
}: EmptyStateProps) {
	const isNoResults = variant === "no-results";

	return (
		<AnimateIn>
			<div
				className={cn(
					"flex flex-col items-center justify-center gap-6 rounded-xl border border-dashed border-border text-center",
					isNoResults
						? "min-h-[200px] bg-muted/30 py-8 px-6"
						: "min-h-[400px] py-16 px-8",
				)}
			>
				<div
					className={cn(
						"flex items-center justify-center rounded-2xl",
						isNoResults
							? "size-10 bg-muted/50"
							: "size-14 bg-muted/50",
					)}
				>
					<Icon
						className={cn(
							"text-muted-foreground",
							isNoResults ? "size-5" : "size-6",
						)}
					/>
				</div>
				<div className="space-y-1.5">
					<h3
						className={cn(
							"font-semibold tracking-tight",
							isNoResults ? "text-base" : "text-lg",
						)}
					>
						{title}
					</h3>
					<p className="max-w-[320px] text-sm text-muted-foreground">
						{description}
					</p>
				</div>
				{!isNoResults && (actionLabel || secondaryAction) && (
					<div className="flex items-center gap-3">
						{actionLabel && onAction && (
							<Button onClick={onAction}>{actionLabel}</Button>
						)}
						{secondaryAction && (
							<Button
								variant="ghost"
								onClick={secondaryAction.onClick}
							>
								{secondaryAction.label}
							</Button>
						)}
					</div>
				)}
			</div>
		</AnimateIn>
	);
}
