import { FormSkeleton } from "@/components/shared/loading-skeleton";

export default function Loading() {
	return (
		<div className="space-y-6">
			<div className="space-y-1">
				<div className="h-7 w-40 animate-pulse rounded-lg bg-muted" />
				<div className="h-4 w-56 animate-pulse rounded bg-muted" />
			</div>
			<FormSkeleton />
		</div>
	);
}
