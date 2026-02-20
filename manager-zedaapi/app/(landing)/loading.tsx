import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
	return (
		<div className="flex flex-col">
			{/* Hero skeleton */}
			<div className="flex flex-col items-center gap-6 px-4 pb-16 pt-20 sm:pb-24 sm:pt-28">
				<Skeleton className="h-8 w-48 rounded-full" />
				<Skeleton className="h-12 w-full max-w-lg" />
				<Skeleton className="h-5 w-full max-w-md" />
				<div className="flex gap-4">
					<Skeleton className="h-11 w-36 rounded-lg" />
					<Skeleton className="h-11 w-40 rounded-lg" />
				</div>
				<Skeleton className="h-64 w-full max-w-2xl rounded-xl" />
			</div>

			{/* Stats skeleton */}
			<div className="border-y border-border bg-muted/30">
				<div className="mx-auto grid max-w-7xl grid-cols-2 divide-x divide-border px-4 lg:grid-cols-4">
					{Array.from({ length: 4 }).map((_, i) => (
						<div
							key={i}
							className="flex flex-col items-center gap-2 px-4 py-6"
						>
							<Skeleton className="h-9 w-24" />
							<Skeleton className="h-4 w-28" />
						</div>
					))}
				</div>
			</div>

			{/* Features skeleton */}
			<div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 sm:py-24 lg:px-8">
				<div className="flex flex-col items-center gap-4">
					<Skeleton className="h-9 w-80" />
					<Skeleton className="h-5 w-64" />
				</div>
				<div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{Array.from({ length: 6 }).map((_, i) => (
						<Skeleton key={i} className="h-32 w-full rounded-2xl" />
					))}
				</div>
			</div>
		</div>
	);
}
