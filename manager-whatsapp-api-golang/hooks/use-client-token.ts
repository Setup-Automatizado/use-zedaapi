/**
 * useClientToken Hook
 *
 * Fetches the global client authentication token from the config API.
 *
 * @example
 * ```tsx
 * const { clientToken, isLoading } = useClientToken();
 * ```
 */

import useSWR from "swr";

interface ConfigResponse {
	clientToken: string;
}

const fetcher = async (url: string): Promise<ConfigResponse> => {
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error("Failed to fetch config");
	}
	return response.json();
};

export function useClientToken() {
	const { data, error, isLoading } = useSWR<ConfigResponse>(
		"/api/config",
		fetcher,
		{
			revalidateOnFocus: false,
			revalidateOnReconnect: false,
			dedupingInterval: 60000, // Cache for 1 minute
		},
	);

	return {
		clientToken: data?.clientToken || "",
		isLoading,
		isError: Boolean(error),
	};
}
