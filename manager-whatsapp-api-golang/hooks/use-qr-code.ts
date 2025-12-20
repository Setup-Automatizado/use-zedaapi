/**
 * QR Code Hook with Auto-Refresh
 *
 * Fetches QR code image for WhatsApp pairing with automatic refresh.
 * Polls every 30 seconds while not connected, stops when connected.
 *
 * @example
 * ```tsx
 * const { image, isConnected, isLoading, error, refresh } = useQRCode(instanceId, { enabled: true });
 * ```
 */

"use client";

import { useEffect } from "react";
import type { QRCodeResponse } from "@/types";
import { useInstanceStatus } from "./use-instance-status";
import { usePolling } from "./use-polling";

/**
 * QR code hook options
 */
export interface UseQRCodeOptions {
	/**
	 * Polling interval in milliseconds
	 * @default 30000 (30 seconds)
	 */
	interval?: number;

	/**
	 * Enable or disable automatic polling
	 * When true, polls only while not connected
	 * @default true
	 */
	autoPoll?: boolean;

	/**
	 * Enable or disable the hook entirely
	 * When false, no requests are made until enabled
	 * @default true
	 */
	enabled?: boolean;
}

/**
 * QR code hook result
 */
export interface UseQRCodeResult {
	/** Base64-encoded QR code image with data URI prefix */
	image: string | undefined;

	/** Connection status from instance status endpoint */
	isConnected: boolean;

	/** Loading state */
	isLoading: boolean;

	/** Error message */
	error: string | undefined;

	/** Revalidation state */
	isRefreshing: boolean;

	/** Manual refresh function */
	refresh: () => Promise<void>;
}

/**
 * Hook to fetch QR code with automatic refresh
 *
 * Features:
 * - Auto-refreshes every 30 seconds by default
 * - Stops polling when instance is connected
 * - Integrates with instance status
 * - Error handling and retry logic
 * - Only fetches when enabled (user must request QR code first)
 *
 * @param instanceId - Instance ID (UUID)
 * @param options - QR code configuration options
 * @returns QR code image and connection status
 */
export function useQRCode(
	instanceId: string | null | undefined,
	options: UseQRCodeOptions = {},
): UseQRCodeResult {
	const { interval = 30000, autoPoll = true, enabled = true } = options;

	// Monitor connection status only when enabled
	const { isConnected, isLoading: statusLoading } = useInstanceStatus(
		instanceId,
		{
			enabled: enabled && autoPoll,
			interval: 5000,
		},
	);

	// Determine if polling should be enabled
	// Poll only when enabled AND auto-poll is enabled AND instance is not connected
	const shouldPoll = enabled && autoPoll && !isConnected;

	const endpoint =
		instanceId && enabled ? `/api/instances/${instanceId}/qr-code` : null;

	const {
		data,
		error: requestError,
		isLoading: qrLoading,
		isValidating,
		mutate,
	} = usePolling<QRCodeResponse>(endpoint, {
		interval,
		enabled: shouldPoll,
		dedupingInterval: 2000,
	});

	// Stop polling when connected
	useEffect(() => {
		if (isConnected && autoPoll) {
			// Clear the QR code when connected
			mutate();
		}
	}, [isConnected, autoPoll, mutate]);

	return {
		image: data?.image,
		isConnected,
		isLoading: enabled ? qrLoading || statusLoading : false,
		error: requestError?.message,
		isRefreshing: isValidating,
		refresh: async () => {
			await mutate();
		},
	};
}
