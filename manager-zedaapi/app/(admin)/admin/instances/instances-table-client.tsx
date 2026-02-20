"use client";

import { useState, useCallback, useEffect } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Smartphone } from "lucide-react";
import { cn } from "@/lib/utils";
import { getAdminInstances } from "@/server/actions/admin";

interface AdminInstance {
	id: string;
	name: string;
	userName: string;
	status: string;
	phone: string | null;
	whatsappConnected: boolean;
	planName: string | null;
	createdAt: Date;
}

const statusConfig: Record<string, { label: string; className: string }> = {
	connected: { label: "Conectado", className: "bg-primary/10 text-primary" },
	connecting: {
		label: "Conectando",
		className: "bg-chart-2/10 text-chart-2",
	},
	disconnected: {
		label: "Desconectado",
		className: "bg-muted text-muted-foreground",
	},
	creating: {
		label: "Criando",
		className: "bg-chart-2/10 text-chart-2",
	},
	error: { label: "Erro", className: "bg-destructive/10 text-destructive" },
	banned: {
		label: "Banido",
		className: "bg-destructive/10 text-destructive",
	},
};

interface InstancesTableClientProps {
	initialData: AdminInstance[];
	initialTotal: number;
}

export function InstancesTableClient({
	initialData,
	initialTotal,
}: InstancesTableClientProps) {
	const [data, setData] = useState<AdminInstance[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);

	const load = useCallback(async () => {
		setLoading(true);
		const res = await getAdminInstances(page);
		if (res.success && res.data) {
			setData(res.data.items);
			setTotal(res.data.total);
		}
		setLoading(false);
	}, [page]);

	useEffect(() => {
		if (page > 1) {
			load();
		}
	}, [load, page]);

	const columns: Column<AdminInstance>[] = [
		{
			key: "name",
			header: "Instancia",
			cell: (row) => <span className="font-medium">{row.name}</span>,
		},
		{
			key: "user",
			header: "Usuario",
			cell: (row) => <span className="text-sm">{row.userName}</span>,
		},
		{
			key: "phone",
			header: "Telefone",
			cell: (row) => (
				<span className="text-muted-foreground">
					{row.phone ?? "-"}
				</span>
			),
		},
		{
			key: "plan",
			header: "Plano",
			cell: (row) =>
				row.planName ? (
					<Badge variant="outline">{row.planName}</Badge>
				) : (
					<span className="text-xs text-muted-foreground">-</span>
				),
		},
		{
			key: "status",
			header: "Status",
			cell: (row) => {
				const config = statusConfig[row.status] ?? {
					label: row.status,
					className: "bg-muted text-muted-foreground",
				};
				return (
					<Badge variant="secondary" className={cn(config.className)}>
						{config.label}
					</Badge>
				);
			},
		},
		{
			key: "createdAt",
			header: "Criado em",
			cell: (row) => new Date(row.createdAt).toLocaleDateString("pt-BR"),
		},
	];

	return (
		<DataTable
			columns={columns}
			data={data}
			loading={loading}
			emptyIcon={Smartphone}
			emptyTitle="Nenhuma instancia"
			emptyDescription="Nenhuma instancia encontrada."
			searchPlaceholder="Buscar instancias..."
			page={page}
			pageSize={20}
			totalCount={total}
			onPageChange={setPage}
		/>
	);
}
