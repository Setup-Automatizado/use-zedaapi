import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<Skeleton className="h-9 w-64" />
				<Skeleton className="h-9 w-32" />
			</div>
			<div className="rounded-xl border">
				<div className="border-b px-4 py-3">
					<div className="flex gap-4">
						{[1, 2, 3, 4].map((i) => (
							<Skeleton key={i} className="h-4 flex-1" />
						))}
					</div>
				</div>
				{Array.from({ length: rows }).map((_, i) => (
					<div key={i} className="border-b px-4 py-3 last:border-0">
						<div className="flex gap-4">
							{[1, 2, 3, 4].map((j) => (
								<Skeleton key={j} className="h-4 flex-1" />
							))}
						</div>
					</div>
				))}
			</div>
			<div className="flex items-center justify-between">
				<Skeleton className="h-4 w-48" />
				<div className="flex gap-2">
					<Skeleton className="h-9 w-9" />
					<Skeleton className="h-9 w-9" />
				</div>
			</div>
		</div>
	);
}

export function CardSkeleton() {
	return (
		<Card>
			<CardHeader>
				<Skeleton className="h-4 w-32" />
				<Skeleton className="h-3 w-48" />
			</CardHeader>
			<CardContent>
				<Skeleton className="h-8 w-24" />
			</CardContent>
		</Card>
	);
}

export function CardsSkeleton({ count = 4 }: { count?: number }) {
	return (
		<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{Array.from({ length: count }).map((_, i) => (
				<CardSkeleton key={i} />
			))}
		</div>
	);
}

export function FormSkeleton({ fields = 4 }: { fields?: number }) {
	return (
		<div className="space-y-6">
			{Array.from({ length: fields }).map((_, i) => (
				<div key={i} className="space-y-2">
					<Skeleton className="h-4 w-24" />
					<Skeleton className="h-10 w-full" />
				</div>
			))}
			<Skeleton className="h-10 w-32" />
		</div>
	);
}

export function GridSkeleton({ count = 3 }: { count?: number }) {
	return (
		<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{Array.from({ length: count }).map((_, i) => (
				<Skeleton key={i} className="h-[72px] rounded-xl" />
			))}
		</div>
	);
}

export function ProfileFormSkeleton() {
	return (
		<Card>
			<CardHeader>
				<Skeleton className="h-5 w-48" />
				<Skeleton className="h-3 w-64" />
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="grid gap-4 sm:grid-cols-2">
					{[1, 2, 3, 4].map((i) => (
						<div key={i} className="space-y-2">
							<Skeleton className="h-4 w-24" />
							<Skeleton className="h-10 w-full" />
						</div>
					))}
				</div>
				<Skeleton className="ml-auto h-10 w-32" />
			</CardContent>
		</Card>
	);
}

export function DetailSkeleton() {
	return (
		<div className="space-y-6">
			<div className="flex items-center gap-3">
				<Skeleton className="size-8 rounded-lg" />
				<div className="flex-1 space-y-1">
					<Skeleton className="h-7 w-48" />
					<Skeleton className="h-4 w-32" />
				</div>
				<div className="flex gap-2">
					<Skeleton className="h-9 w-24" />
					<Skeleton className="h-9 w-24" />
				</div>
			</div>
			<div className="grid gap-4 lg:grid-cols-2">
				{[1, 2, 3].map((i) => (
					<Card key={i} className={i === 3 ? "lg:col-span-2" : ""}>
						<CardHeader>
							<Skeleton className="h-5 w-40" />
						</CardHeader>
						<CardContent className="space-y-3">
							{[1, 2, 3].map((j) => (
								<Skeleton key={j} className="h-4 w-full" />
							))}
						</CardContent>
					</Card>
				))}
			</div>
		</div>
	);
}

export function ChartSkeleton() {
	return (
		<Card>
			<CardHeader>
				<Skeleton className="h-5 w-40" />
				<Skeleton className="h-3 w-56" />
			</CardHeader>
			<CardContent>
				<Skeleton className="h-64 w-full" />
			</CardContent>
		</Card>
	);
}
