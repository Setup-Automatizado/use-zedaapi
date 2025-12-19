/**
 * Instance Status Hook with Polling
 *
 * Fetches instance connection status with automatic polling every 5 seconds.
 * Polling can be enabled or disabled dynamically.
 *
 * @example
 * ```tsx
 * const { status, isConnected, smartphoneConnected, error, isLoading, refresh } =
 *   useInstanceStatus(instanceId, { enabled: true });
 * ```
 */

'use client';

import { usePolling } from './use-polling';
import type { InstanceStatus } from '@/types';

/**
 * Instance status hook options
 */
export interface UseInstanceStatusOptions {
  /**
   * Enable or disable polling
   * @default true
   */
  enabled?: boolean;

  /**
   * Polling interval in milliseconds
   * @default 5000 (5 seconds)
   */
  interval?: number;
}

/**
 * Instance status hook result
 */
export interface UseInstanceStatusResult {
  /** Complete status object */
  status: InstanceStatus | undefined;

  /** Overall connection status */
  isConnected: boolean;

  /** Physical device connection status */
  smartphoneConnected: boolean;

  /** Error message from status or request error */
  error: string | undefined;

  /** Loading state */
  isLoading: boolean;

  /** Revalidation state */
  isValidating: boolean;

  /** Manual refresh function */
  refresh: () => Promise<void>;
}

/**
 * Hook to fetch instance status with automatic polling
 *
 * Features:
 * - Polls every 5 seconds by default
 * - Can be enabled/disabled dynamically
 * - Automatic error handling
 * - Focus revalidation
 * - Network error recovery
 *
 * @param instanceId - Instance ID (UUID)
 * @param options - Polling configuration options
 * @returns Instance status with connection information
 */
export function useInstanceStatus(
  instanceId: string | null | undefined,
  options: UseInstanceStatusOptions = {}
): UseInstanceStatusResult {
  const { enabled = true, interval = 5000 } = options;

  const endpoint = instanceId ? `/api/instances/${instanceId}/status` : null;

  const { data, error: requestError, isLoading, isValidating, mutate } = usePolling<InstanceStatus>(
    endpoint,
    {
      interval,
      enabled,
      dedupingInterval: 1000,
    }
  );

  return {
    status: data,
    isConnected: data?.connected ?? false,
    smartphoneConnected: data?.smartphoneConnected ?? false,
    error: data?.error || requestError?.message,
    isLoading,
    isValidating,
    refresh: async () => {
      await mutate();
    },
  };
}
