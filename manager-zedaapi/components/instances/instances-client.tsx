"use client";

import { useState } from "react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { EmptyState } from "@/components/shared/empty-state";
import { StatusBadge } from "@/components/shared/status-badge";
import { Smartphone, Search } from "lucide-react";

interface Instance {
	id: string;
	name: string;
	status: string;
	phone: string | null;
	createdAt: Date;
}

interface InstancesClientProps {
	instances: Instance[];
}

export function InstancesClient({ instances }: InstancesClientProps) {
	const [search, setSearch] = useState("");

	const filtered = instances.filter(
		(instance) =>
			instance.name.toLowerCase().includes(search.toLowerCase()) ||
			instance.phone?.includes(search),
	);

	if (instances.length === 0) {
		return (
			<EmptyState
				icon={Smartphone}
				title="Nenhuma instância"
				description="Crie sua primeira instância WhatsApp para começar."
				actionLabel="Criar Instância"
				onAction={() => {}}
			/>
		);
	}

	return (
		<div className="space-y-4">
			<div className="relative max-w-sm">
				<Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
				<Input
					placeholder="Buscar instâncias..."
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					className="pl-9"
				/>
			</div>

			<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
				{filtered.map((instance) => (
					<Link key={instance.id} href={`/instancias/${instance.id}`}>
						<Card className="transition-colors hover:bg-accent/50 cursor-pointer">
							<CardContent className="flex items-center gap-3 py-4">
								<div className="flex size-10 items-center justify-center rounded-lg bg-muted">
									<Smartphone className="size-5 text-muted-foreground" />
								</div>
								<div className="flex-1 min-w-0">
									<p className="truncate text-sm font-medium">
										{instance.name}
									</p>
									<p className="text-xs text-muted-foreground">
										{instance.phone ?? "Sem telefone"}
									</p>
								</div>
								<StatusBadge
									status={instance.status}
									type="instance"
								/>
							</CardContent>
						</Card>
					</Link>
				))}
			</div>

			{filtered.length === 0 && search && (
				<p className="text-center text-sm text-muted-foreground py-8">
					Nenhuma instância encontrada para &ldquo;{search}&rdquo;
				</p>
			)}
		</div>
	);
}
