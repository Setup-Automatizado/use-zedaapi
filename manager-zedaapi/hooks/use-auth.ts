"use client";

import { authClient } from "@/lib/auth-client";

export function useAuth() {
	const session = authClient.useSession();

	return {
		session: session.data,
		user: session.data?.user ?? null,
		isLoading: session.isPending,
		isRefetching: session.isRefetching,
		error: session.error,
		isAuthenticated: !!session.data?.user,
		isAdmin: session.data?.user?.role === "admin",
		refetch: session.refetch,
	};
}
