import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';

/**
 * Loading state for the webhooks configuration page.
 * Displays skeleton placeholders while data is being fetched.
 */
export default function WebhooksLoading() {
  return (
    <div className="space-y-6">
      {/* Header Skeleton */}
      <div className="flex items-center gap-4">
        <Skeleton className="h-10 w-10" />
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
        </div>
      </div>

      {/* Card Skeleton */}
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-full max-w-2xl" />
        </CardHeader>
        <CardContent className="space-y-6">
          {/* 7 Webhook Fields */}
          {Array.from({ length: 7 }).map((_, i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-4 w-full max-w-lg" />
              <Skeleton className="h-10 w-full" />
            </div>
          ))}

          {/* Separator */}
          <Skeleton className="h-px w-full" />

          {/* Toggle */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-2">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-4 w-96" />
            </div>
            <Skeleton className="h-6 w-11 rounded-full" />
          </div>

          {/* Buttons */}
          <div className="flex justify-between">
            <Skeleton className="h-10 w-32" />
            <div className="flex gap-3">
              <Skeleton className="h-10 w-24" />
              <Skeleton className="h-10 w-32" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
