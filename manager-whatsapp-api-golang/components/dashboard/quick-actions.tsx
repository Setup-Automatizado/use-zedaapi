import { List, Plus } from "lucide-react";
import Link from "next/link";
import type * as React from "react";
import { Button } from "@/components/ui/button";

interface QuickAction {
	label: string;
	icon: React.ComponentType<{ className?: string }>;
	href: string;
	variant: "default" | "outline" | "secondary";
}

const actions: QuickAction[] = [
	{
		label: "New Instance",
		icon: Plus,
		href: "/instances/new",
		variant: "default",
	},
	{
		label: "View All",
		icon: List,
		href: "/instances",
		variant: "outline",
	},
];

export function QuickActions() {
	return (
		<div className="flex items-center gap-2">
			{actions.map((action) => (
				<Button key={action.href} variant={action.variant} size="sm" asChild>
					<Link href={action.href}>
						<action.icon className="h-4 w-4" data-icon="inline-start" />
						{action.label}
					</Link>
				</Button>
			))}
		</div>
	);
}
