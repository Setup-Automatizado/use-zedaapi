import { CardsSkeleton, GridSkeleton } from "@/components/shared/loading-skeleton";

export default function Loading() {
	return (
		<div className="space-y-6">
			<div className="space-y-1">
				<div className="h-7 w-32 animate-pulse rounded-lg bg-muted" />
				<div className="h-4 w-48 animate-pulse rounded bg-muted" />
			</div>
			<CardsSkeleton count={4} />
			<GridSkeleton count={3} />
		</div>
	);
}
