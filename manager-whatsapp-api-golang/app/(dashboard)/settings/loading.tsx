/**
 * Settings Page Loading State
 *
 * Skeleton UI displayed while settings sections are being loaded.
 */

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export default function SettingsLoading() {
	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="space-y-2">
				<Skeleton className="h-9 w-32" />
				<Skeleton className="h-5 w-72" />
			</div>

			{/* Settings Cards Grid */}
			<div className="grid gap-4 md:grid-cols-2">
				{Array.from({ length: 3 }).map((_, i) => (
					<Card key={i}>
						<CardHeader>
							<div className="flex items-center gap-3">
								<Skeleton className="h-10 w-10 rounded-full" />
								<div className="space-y-2">
									<Skeleton className="h-5 w-24" />
									<Skeleton className="h-4 w-48" />
								</div>
							</div>
						</CardHeader>
						<CardContent>
							<Skeleton className="h-10 w-24" />
						</CardContent>
					</Card>
				))}
			</div>
		</div>
	);
}
