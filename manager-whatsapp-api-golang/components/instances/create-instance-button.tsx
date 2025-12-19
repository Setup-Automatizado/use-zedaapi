import * as React from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

export interface CreateInstanceButtonProps {
	className?: string;
}

export function CreateInstanceButton({ className }: CreateInstanceButtonProps) {
	return (
		<Button asChild className={className}>
			<Link href="/instances/new">
				<Plus className="h-4 w-4" />
				New Instance
			</Link>
		</Button>
	);
}
