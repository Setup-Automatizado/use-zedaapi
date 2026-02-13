"use client";

import { Trash2, Users } from "lucide-react";
import { useTransition } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { PoolGroup } from "@/types/pool";

interface GroupCardProps {
	group: PoolGroup;
	onDelete: (id: string) => Promise<void>;
}

export function GroupCard({ group, onDelete }: GroupCardProps) {
	const [isDeleting, startDelete] = useTransition();

	const handleDelete = () => {
		startDelete(async () => {
			try {
				await onDelete(group.id);
				toast.success("Group deleted");
			} catch {
				toast.error("Failed to delete group");
			}
		});
	};

	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between pb-3">
				<div className="flex items-center gap-2">
					<Users className="h-4 w-4 text-muted-foreground" />
					<CardTitle className="text-base">{group.name}</CardTitle>
				</div>
				<Button
					variant="ghost"
					size="icon"
					className="h-8 w-8 text-destructive"
					onClick={handleDelete}
					disabled={isDeleting}
				>
					<Trash2 className="h-4 w-4" />
				</Button>
			</CardHeader>
			<CardContent>
				<div className="grid grid-cols-2 gap-4 text-sm">
					<div>
						<p className="text-muted-foreground">Max Instances</p>
						<p className="font-medium">{group.maxInstances}</p>
					</div>
					{group.countryCode && (
						<div>
							<p className="text-muted-foreground">Country</p>
							<Badge variant="outline">{group.countryCode}</Badge>
						</div>
					)}
				</div>
			</CardContent>
		</Card>
	);
}
