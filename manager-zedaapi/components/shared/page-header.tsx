"use client";

import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

interface PageHeaderProps {
	title: string;
	description?: string;
	action?: React.ReactNode;
	backHref?: string;
}

export function PageHeader({
	title,
	description,
	action,
	backHref,
}: PageHeaderProps) {
	return (
		<div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
			<div className="space-y-1">
				<div className="flex items-center gap-2">
					{backHref && (
						<Button
							variant="ghost"
							size="icon-sm"
							asChild
							className="text-muted-foreground"
						>
							<Link href={backHref}>
								<ArrowLeft className="size-4" />
							</Link>
						</Button>
					)}
					<h1 className="text-2xl font-bold tracking-tight">
						{title}
					</h1>
				</div>
				{description && (
					<p className="text-sm text-muted-foreground">
						{description}
					</p>
				)}
			</div>
			{action && <div className="mt-3 sm:mt-0">{action}</div>}
		</div>
	);
}
