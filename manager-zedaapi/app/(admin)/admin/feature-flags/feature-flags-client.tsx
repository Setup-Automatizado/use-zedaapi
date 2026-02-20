"use client";

import { useState, useCallback } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Flag } from "lucide-react";
import { EmptyState } from "@/components/shared/empty-state";
import { toast } from "sonner";
import { getFeatureFlags, toggleFeatureFlag } from "@/server/actions/admin";

interface FeatureFlag {
	id: string;
	key: string;
	enabled: boolean;
	description: string | null;
}

interface FeatureFlagsClientProps {
	initialFlags: FeatureFlag[];
}

export function FeatureFlagsClient({ initialFlags }: FeatureFlagsClientProps) {
	const [flags, setFlags] = useState<FeatureFlag[]>(initialFlags);

	const reload = useCallback(async () => {
		const res = await getFeatureFlags();
		if (res.success && res.data) setFlags(res.data);
	}, []);

	const handleToggle = async (key: string) => {
		const res = await toggleFeatureFlag(key);
		if (res.success) {
			toast.success("Feature flag atualizada");
			reload();
		} else {
			toast.error(res.error ?? "Erro ao atualizar flag");
		}
	};

	if (flags.length === 0) {
		return (
			<EmptyState
				icon={Flag}
				title="Nenhuma feature flag"
				description="Nenhuma feature flag cadastrada no sistema."
			/>
		);
	}

	return (
		<Card>
			<CardContent className="divide-y p-0">
				{flags.map((flag) => (
					<div
						key={flag.id}
						className="flex items-center justify-between px-6 py-4"
					>
						<div className="space-y-1">
							<div className="flex items-center gap-2">
								<Label className="font-mono text-sm">
									{flag.key}
								</Label>
								<Badge
									variant={
										flag.enabled ? "default" : "secondary"
									}
									className="text-[10px]"
								>
									{flag.enabled ? "Ativo" : "Inativo"}
								</Badge>
							</div>
							{flag.description && (
								<p className="text-xs text-muted-foreground">
									{flag.description}
								</p>
							)}
						</div>
						<Switch
							checked={flag.enabled}
							onCheckedChange={() => handleToggle(flag.key)}
						/>
					</div>
				))}
			</CardContent>
		</Card>
	);
}
