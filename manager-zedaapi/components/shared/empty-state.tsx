import type { LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";

interface EmptyStateProps {
	icon: LucideIcon;
	title: string;
	description: string;
	actionLabel?: string;
	onAction?: () => void;
}

export function EmptyState({
	icon: Icon,
	title,
	description,
	actionLabel,
	onAction,
}: EmptyStateProps) {
	return (
		<div className="flex min-h-[400px] flex-col items-center justify-center gap-6 rounded-xl border border-dashed border-border py-16 px-8 text-center">
			<div className="flex size-14 items-center justify-center rounded-2xl bg-muted/50">
				<Icon className="size-6 text-muted-foreground" />
			</div>
			<div className="space-y-1.5">
				<h3 className="text-lg font-semibold tracking-tight">
					{title}
				</h3>
				<p className="max-w-[320px] text-sm text-muted-foreground">
					{description}
				</p>
			</div>
			{actionLabel && onAction && (
				<Button onClick={onAction}>{actionLabel}</Button>
			)}
		</div>
	);
}
