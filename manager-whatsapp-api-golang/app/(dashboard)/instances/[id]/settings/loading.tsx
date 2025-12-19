import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';

/**
 * Loading state for the instance settings page.
 * Displays skeleton placeholders while data is being fetched.
 */
export default function SettingsLoading() {
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
          <Skeleton className="h-6 w-64" />
          <Skeleton className="h-4 w-full max-w-2xl" />
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Auto Read Message Toggle */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-2">
              <Skeleton className="h-5 w-56" />
              <Skeleton className="h-4 w-80" />
            </div>
            <Skeleton className="h-6 w-11 rounded-full" />
          </div>

          {/* Separator */}
          <Skeleton className="h-px w-full" />

          {/* Call Reject Auto Toggle */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-2">
              <Skeleton className="h-5 w-64" />
              <Skeleton className="h-4 w-96" />
            </div>
            <Skeleton className="h-6 w-11 rounded-full" />
          </div>

          {/* Call Reject Message (conditional) */}
          <div className="space-y-2">
            <Skeleton className="h-5 w-56" />
            <Skeleton className="h-4 w-full max-w-lg" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-4 w-16 ml-auto" />
          </div>

          {/* Separator */}
          <Skeleton className="h-px w-full" />

          {/* Notify Sent By Me Toggle */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-2">
              <Skeleton className="h-5 w-52" />
              <Skeleton className="h-4 w-full max-w-md" />
            </div>
            <Skeleton className="h-6 w-11 rounded-full" />
          </div>

          {/* Buttons */}
          <div className="flex justify-end gap-3">
            <Skeleton className="h-10 w-24" />
            <Skeleton className="h-10 w-44" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
