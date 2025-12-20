/**
 * New Instance Page Loading State
 *
 * Skeleton UI displayed while the form is being prepared.
 */

import { PageHeader } from "@/components/shared/page-header";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export default function NewInstanceLoading() {
	return (
		<div className="space-y-6">
			<PageHeader
				title="Nova Instância"
				description="Configure uma nova instância do WhatsApp para começar a enviar e receber mensagens"
			/>

			<Card>
				<CardHeader>
					<Skeleton className="h-6 w-48" />
					<Skeleton className="h-4 w-full max-w-lg" />
				</CardHeader>
				<CardContent className="space-y-6">
					{/* Instance Name Field */}
					<div className="space-y-2">
						<Skeleton className="h-4 w-32" />
						<Skeleton className="h-10 w-full" />
						<Skeleton className="h-3 w-64" />
					</div>

					{/* Webhook Delivery Field */}
					<div className="space-y-2">
						<Skeleton className="h-4 w-40" />
						<Skeleton className="h-10 w-full" />
						<Skeleton className="h-3 w-80" />
					</div>

					{/* Webhook Received Field */}
					<div className="space-y-2">
						<Skeleton className="h-4 w-44" />
						<Skeleton className="h-10 w-full" />
						<Skeleton className="h-3 w-72" />
					</div>

					{/* Separator */}
					<Skeleton className="h-px w-full" />

					{/* Toggles */}
					<div className="flex items-center justify-between rounded-lg border p-4">
						<div className="space-y-2">
							<Skeleton className="h-5 w-48" />
							<Skeleton className="h-4 w-80" />
						</div>
						<Skeleton className="h-6 w-11 rounded-full" />
					</div>

					<div className="flex items-center justify-between rounded-lg border p-4">
						<div className="space-y-2">
							<Skeleton className="h-5 w-56" />
							<Skeleton className="h-4 w-96" />
						</div>
						<Skeleton className="h-6 w-11 rounded-full" />
					</div>

					{/* Buttons */}
					<div className="flex justify-end gap-3 pt-4">
						<Skeleton className="h-10 w-24" />
						<Skeleton className="h-10 w-36" />
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
