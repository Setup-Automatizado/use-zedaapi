import {
	CardsSkeleton,
	ChartSkeleton,
} from "@/components/shared/loading-skeleton";

export default function Loading() {
	return (
		<div className="space-y-6">
			<div className="space-y-1">
				<div className="h-7 w-36 animate-pulse rounded-lg bg-muted" />
				<div className="h-4 w-52 animate-pulse rounded bg-muted" />
			</div>
			<CardsSkeleton count={6} />
			<div className="grid gap-4 lg:grid-cols-2">
				<ChartSkeleton />
				<ChartSkeleton />
			</div>
		</div>
	);
}
