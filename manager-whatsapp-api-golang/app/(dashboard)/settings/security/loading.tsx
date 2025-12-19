/**
 * Security Settings Page Loading State
 *
 * Skeleton UI displayed while security settings are being loaded.
 */

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';

export default function SecurityLoading() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-5 w-80" />
      </div>

      {/* Two-Factor Authentication Section */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-full" />
            <div className="space-y-2">
              <Skeleton className="h-6 w-56" />
              <Skeleton className="h-4 w-72" />
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Status message */}
          <div className="flex items-center gap-2 p-3 rounded-lg border">
            <Skeleton className="h-5 w-5" />
            <Skeleton className="h-4 w-64" />
          </div>

          {/* Method indicator */}
          <div className="flex items-center gap-2 p-3 rounded-lg border">
            <Skeleton className="h-4 w-4" />
            <Skeleton className="h-4 w-40" />
          </div>

          {/* Action buttons */}
          <div className="flex gap-3">
            <Skeleton className="h-10 w-28" />
            <Skeleton className="h-10 w-48" />
          </div>
        </CardContent>
      </Card>

      {/* Password Section */}
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-24" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-10 w-40" />
        </CardContent>
      </Card>
    </div>
  );
}
