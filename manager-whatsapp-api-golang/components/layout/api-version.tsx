"use client";

import { cn } from "@/lib/utils";
import { useHealth } from "@/hooks";

interface ApiVersionProps {
	className?: string;
	collapsed?: boolean;
}

export function ApiVersion({ className, collapsed }: ApiVersionProps) {
	const { health, isLoading } = useHealth({
		interval: 60000, // Check every minute
		includeReadiness: false, // Only need basic health for version
	});

	const version = health?.version;

	if (isLoading || !version) {
		return null;
	}

	if (collapsed) {
		return (
			<div
				className={cn(
					"text-[10px] text-muted-foreground/60 text-center",
					className,
				)}
				title={`API v${version}`}
			>
				v{version.split("-")[0]}
			</div>
		);
	}

	return (
		<div
			className={cn(
				"text-xs text-muted-foreground/60 px-3 text-center",
				className,
			)}
		>
			API v{version}
		</div>
	);
}
