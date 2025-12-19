/**
 * Generic Polling Hook
 *
 * Provides configurable polling functionality using SWR.
 * Supports dynamic interval adjustment and enable/disable control.
 *
 * @example
 * ```tsx
 * const { data, error, isLoading, mutate } = usePolling(
 *   '/api/status',
 *   { interval: 5000, enabled: isConnected }
 * );
 * ```
 */

'use client';

import useSWR, { type SWRConfiguration } from 'swr';

/**
 * Polling hook configuration options
 */
export interface UsePollingOptions<T> extends SWRConfiguration<T> {
  /**
   * Polling interval in milliseconds
   * @default 30000 (30 seconds)
   */
  interval?: number;

  /**
   * Enable or disable polling
   * @default true
   */
  enabled?: boolean;

  /**
   * Custom fetcher function
   * If not provided, uses default fetch with error handling
   */
  fetcher?: (url: string) => Promise<T>;

  /**
   * Dedupe interval in milliseconds
   * Prevents duplicate requests within this time window
   * @default 2000
   */
  dedupingInterval?: number;
}

/**
 * Polling hook result
 */
export interface UsePollingResult<T> {
  /** Response data */
  data: T | undefined;

  /** Error object if request failed */
  error: Error | undefined;

  /** Loading state (true on initial load) */
  isLoading: boolean;

  /** Validating state (true on revalidation) */
  isValidating: boolean;

  /** Manual revalidation function */
  mutate: () => Promise<T | undefined>;
}

/**
 * Default fetcher with error handling
 */
async function defaultFetcher<T>(url: string): Promise<T> {
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    cache: 'no-store',
  });

  if (!response.ok) {
    const error = new Error(`HTTP ${response.status}: ${response.statusText}`) as Error & { status: number };
    error.status = response.status;
    throw error;
  }

  return response.json();
}

/**
 * Generic polling hook with configurable interval
 *
 * Features:
 * - Automatic polling at specified interval
 * - Enable/disable polling dynamically
 * - Error handling and retry logic
 * - Deduplication to prevent duplicate requests
 * - Focus revalidation
 * - Network error recovery
 *
 * @param key - SWR cache key (usually the API endpoint)
 * @param options - Polling configuration options
 * @returns Polling result with data, error, loading states, and mutate function
 */
export function usePolling<T = unknown>(
  key: string | null,
  options: UsePollingOptions<T> = {}
): UsePollingResult<T> {
  const {
    interval = 30000,
    enabled = true,
    fetcher = defaultFetcher,
    dedupingInterval = 2000,
    ...swrOptions
  } = options;

  const { data, error, isLoading, isValidating, mutate } = useSWR<T>(
    enabled && key ? key : null,
    fetcher,
    {
      // Polling configuration
      refreshInterval: enabled ? interval : 0,

      // Deduplication
      dedupingInterval,

      // Revalidation behavior
      revalidateOnFocus: true,
      revalidateOnReconnect: true,
      revalidateIfStale: true,

      // Error handling
      shouldRetryOnError: true,
      errorRetryCount: 3,
      errorRetryInterval: 5000,

      // Performance
      keepPreviousData: true,

      // Merge with custom options
      ...swrOptions,
    }
  );

  return {
    data,
    error,
    isLoading,
    isValidating,
    mutate,
  };
}
