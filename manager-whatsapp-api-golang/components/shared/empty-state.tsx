import type * as React from "react";
import { cn } from "@/lib/utils";

export interface EmptyStateProps {
	icon?: React.ReactNode;
	title: string;
	description?: string;
	action?: React.ReactNode;
	className?: string;
}

export function EmptyState({
	icon,
	title,
	description,
	action,
	className,
}: EmptyStateProps) {
	return (
		<div
			className={cn(
				"flex min-h-[400px] flex-col items-center justify-center rounded-lg border border-dashed p-8 text-center",
				className,
			)}
		>
			{icon && (
				<div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
					{icon}
				</div>
			)}
			<h3 className="mb-2 text-lg font-semibold">{title}</h3>
			{description && (
				<p className="mb-6 max-w-md text-sm text-muted-foreground">
					{description}
				</p>
			)}
			{action && <div>{action}</div>}
		</div>
	);
}
