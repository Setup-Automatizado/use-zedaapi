/**
 * Admin Settings Page Loading State
 *
 * Skeleton UI displayed while admin data is being fetched.
 */

import { Skeleton } from '@/components/ui/skeleton';
import { Card, CardContent, CardHeader } from '@/components/ui/card';

export default function AdminLoading() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <Skeleton className="h-9 w-48" />
          <Skeleton className="h-5 w-72" />
        </div>
        <Skeleton className="h-10 w-32" />
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-4" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-12" />
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Users Table */}
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border">
            {/* Table Header */}
            <div className="border-b bg-muted/50 p-4">
              <div className="flex items-center gap-4">
                <Skeleton className="h-4 w-32 flex-1" />
                <Skeleton className="h-4 w-24 flex-1" />
                <Skeleton className="h-4 w-16 flex-1" />
                <Skeleton className="h-4 w-20 flex-1" />
                <Skeleton className="h-4 w-28 flex-1" />
                <Skeleton className="h-4 w-16" />
              </div>
            </div>

            {/* Table Rows */}
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="flex items-center gap-4 border-b p-4 last:border-0"
              >
                <Skeleton className="h-4 w-48 flex-1" />
                <Skeleton className="h-4 w-24 flex-1" />
                <Skeleton className="h-6 w-16 rounded-full flex-1" />
                <Skeleton className="h-6 w-20 rounded-full flex-1" />
                <Skeleton className="h-4 w-32 flex-1" />
                <Skeleton className="h-8 w-8" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
