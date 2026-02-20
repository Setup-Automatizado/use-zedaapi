"use client";

import { useState, useRef } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Activity } from "lucide-react";
import { cn } from "@/lib/utils";
import { getAdminActivityLog } from "@/server/actions/admin";

interface ActivityEntry {
	id: string;
	action: string;
	resource: string;
	resourceId: string;
	userName: string;
	userEmail: string;
	timestamp: string;
	details: string | null;
}

const actionConfig: Record<string, { label: string; className: string }> = {
	create: { label: "Criou", className: "bg-primary/10 text-primary" },
	update: { label: "Atualizou", className: "bg-chart-2/10 text-chart-2" },
	delete: {
		label: "Excluiu",
		className: "bg-destructive/10 text-destructive",
	},
	login: { label: "Login", className: "bg-muted text-muted-foreground" },
	ban: {
		label: "Baniu",
		className: "bg-destructive/10 text-destructive",
	},
	approve: { label: "Aprovou", className: "bg-primary/10 text-primary" },
	"user.banned": {
		label: "Baniu",
		className: "bg-destructive/10 text-destructive",
	},
	"user.unbanned": {
		label: "Desbaniu",
		className: "bg-primary/10 text-primary",
	},
	"user.role_changed": {
		label: "Funcao",
		className: "bg-chart-2/10 text-chart-2",
	},
	"setting.updated": {
		label: "Config",
		className: "bg-chart-2/10 text-chart-2",
	},
	"feature_flag.toggled": {
		label: "Flag",
		className: "bg-chart-2/10 text-chart-2",
	},
};

interface ActivityLogClientProps {
	initialData: ActivityEntry[];
	initialTotal: number;
}

export function ActivityLogClient({
	initialData,
	initialTotal,
}: ActivityLogClientProps) {
	const [data, setData] = useState<ActivityEntry[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);
	const [search, setSearch] = useState("");
	const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

	async function fetchData(pageNum: number, searchTerm?: string) {
		setLoading(true);
		const res = await getAdminActivityLog(pageNum, searchTerm || undefined);
		if (res.success && res.data) {
			setData(res.data.items);
			setTotal(res.data.total);
		}
		setLoading(false);
	}

	function handlePageChange(newPage: number) {
		setPage(newPage);
		fetchData(newPage, search);
	}

	function handleSearch(value: string) {
		setSearch(value);
		clearTimeout(searchTimeoutRef.current);
		searchTimeoutRef.current = setTimeout(() => {
			setPage(1);
			fetchData(1, value);
		}, 300);
	}

	const columns: Column<ActivityEntry>[] = [
		{
			key: "action",
			header: "Acao",
			cell: (row) => {
				const config = actionConfig[row.action] ?? {
					label: row.action,
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
			key: "details",
			header: "Detalhes",
			cell: (row) => (
				<div>
					<p className="text-sm">
						{row.details ?? `${row.action} ${row.resource}`}
					</p>
					<p className="font-mono text-[10px] text-muted-foreground">
						{row.resourceId}
					</p>
				</div>
			),
		},
		{
			key: "user",
			header: "Usuario",
			cell: (row) => (
				<div>
					<p className="text-sm">{row.userName}</p>
					<p className="text-xs text-muted-foreground">
						{row.userEmail}
					</p>
				</div>
			),
		},
		{
			key: "timestamp",
			header: "Data/Hora",
			cell: (row) =>
				new Date(row.timestamp).toLocaleString("pt-BR", {
					day: "2-digit",
					month: "2-digit",
					year: "numeric",
					hour: "2-digit",
					minute: "2-digit",
				}),
		},
	];

	return (
		<DataTable
			columns={columns}
			data={data}
			loading={loading}
			emptyIcon={Activity}
			emptyTitle="Nenhuma atividade"
			emptyDescription="Nenhuma atividade registrada."
			onSearch={handleSearch}
			searchPlaceholder="Buscar no log..."
			page={page}
			pageSize={20}
			totalCount={total}
			onPageChange={handlePageChange}
		/>
	);
}
