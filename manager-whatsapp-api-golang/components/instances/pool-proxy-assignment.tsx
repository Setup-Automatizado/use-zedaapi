"use client";

import { Globe, Loader2, Shield, Unlink } from "lucide-react";
import { useCallback, useEffect, useState, useTransition } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import type { PoolAssignment } from "@/types/pool";
import { sanitizePoolProxyUrl } from "@/types/pool";

export interface PoolProxyAssignmentProps {
	instanceId: string;
	instanceToken: string;
	onUpdate?: () => void;
}

export function PoolProxyAssignment({
	instanceId,
	instanceToken,
	onUpdate,
}: PoolProxyAssignmentProps) {
	const [assignment, setAssignment] = useState<PoolAssignment | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const [isAssigning, startAssign] = useTransition();
	const [isReleasing, startRelease] = useTransition();

	const fetchAssignment = useCallback(async () => {
		try {
			const { fetchInstancePoolAssignment } =
				await import("@/actions/pool");
			const result = await fetchInstancePoolAssignment(
				instanceId,
				instanceToken,
			);
			if (result.success && result.data) {
				setAssignment(result.data);
			} else {
				setAssignment(null);
			}
		} catch {
			setAssignment(null);
		} finally {
			setIsLoading(false);
		}
	}, [instanceId, instanceToken]);

	useEffect(() => {
		fetchAssignment();
	}, [fetchAssignment]);

	const handleAssign = () => {
		startAssign(async () => {
			try {
				const { assignInstancePoolProxy } =
					await import("@/actions/pool");
				const result = await assignInstancePoolProxy(
					instanceId,
					instanceToken,
					{},
				);
				if (result.success && result.data) {
					setAssignment(result.data);
					toast.success("Pool proxy assigned");
					onUpdate?.();
				} else {
					toast.error(result.error || "Failed to assign proxy");
				}
			} catch {
				toast.error("Failed to assign pool proxy");
			}
		});
	};

	const handleRelease = () => {
		startRelease(async () => {
			try {
				const { releaseInstancePoolProxy } =
					await import("@/actions/pool");
				const result = await releaseInstancePoolProxy(
					instanceId,
					instanceToken,
				);
				if (result.success) {
					setAssignment(null);
					toast.success("Pool proxy released");
					onUpdate?.();
				} else {
					toast.error(result.error || "Failed to release proxy");
				}
			} catch {
				toast.error("Failed to release pool proxy");
			}
		});
	};

	if (isLoading) {
		return (
			<Card>
				<CardHeader>
					<Skeleton className="h-5 w-40" />
				</CardHeader>
				<CardContent>
					<Skeleton className="h-20 w-full" />
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between">
				<div className="flex items-center gap-2">
					<Globe className="h-4 w-4 text-muted-foreground" />
					<CardTitle className="text-base">
						Pool Proxy Assignment
					</CardTitle>
				</div>
				{assignment ? (
					<Badge variant="default">Assigned</Badge>
				) : (
					<Badge variant="outline">No Assignment</Badge>
				)}
			</CardHeader>
			<CardContent>
				{assignment ? (
					<div className="space-y-4">
						{assignment.proxyUrl && (
							<div className="flex items-center gap-2 rounded-md bg-muted/50 px-3 py-2">
								<Shield className="h-4 w-4 text-primary shrink-0" />
								<code
									className="text-xs font-mono truncate"
									title={sanitizePoolProxyUrl(
										assignment.proxyUrl,
									)}
								>
									{sanitizePoolProxyUrl(assignment.proxyUrl)}
								</code>
							</div>
						)}
						<div className="grid grid-cols-2 gap-4 text-sm">
							<div>
								<p className="text-muted-foreground">
									Assigned By
								</p>
								<p className="font-medium">
									{assignment.assignedBy}
								</p>
							</div>
							<div>
								<p className="text-muted-foreground">
									Assigned At
								</p>
								<p className="font-medium">
									{new Date(
										assignment.assignedAt,
									).toLocaleString()}
								</p>
							</div>
						</div>
						<Button
							variant="destructive"
							size="sm"
							onClick={handleRelease}
							disabled={isReleasing}
						>
							{isReleasing ? (
								<Loader2 className="mr-2 h-4 w-4 animate-spin" />
							) : (
								<Unlink className="mr-2 h-4 w-4" />
							)}
							Release Pool Proxy
						</Button>
					</div>
				) : (
					<div className="space-y-3">
						<p className="text-sm text-muted-foreground">
							No pool proxy assigned. Assign one automatically
							from the proxy pool.
						</p>
						<Button
							size="sm"
							onClick={handleAssign}
							disabled={isAssigning}
						>
							{isAssigning ? (
								<Loader2 className="mr-2 h-4 w-4 animate-spin" />
							) : (
								<Globe className="mr-2 h-4 w-4" />
							)}
							Assign from Pool
						</Button>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
