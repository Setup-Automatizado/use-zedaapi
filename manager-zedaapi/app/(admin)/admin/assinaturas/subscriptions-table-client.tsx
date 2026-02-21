"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { CreditCard } from "lucide-react";
import { cn } from "@/lib/utils";
import { getAdminSubscriptions } from "@/server/actions/admin";

interface Subscription {
	id: string;
	userName: string;
	userEmail: string;
	planName: string;
	status: string;
	paymentMethod: string;
	currentPeriodEnd: Date;
}

const statusConfig: Record<string, { label: string; className: string }> = {
	active: { label: "Ativa", className: "bg-primary/10 text-primary" },
	trialing: { label: "Trial", className: "bg-chart-2/10 text-chart-2" },
	past_due: {
		label: "Vencida",
		className: "bg-destructive/10 text-destructive",
	},
	canceled: {
		label: "Cancelada",
		className: "bg-muted text-muted-foreground",
	},
	paused: { label: "Pausada", className: "bg-chart-2/10 text-chart-2" },
};

const paymentLabels: Record<string, string> = {
	stripe: "Cartao",
	pix: "PIX",
	boleto: "Boleto",
};

interface SubscriptionsTableClientProps {
	initialData: Subscription[];
	initialTotal: number;
}

export function SubscriptionsTableClient({
	initialData,
	initialTotal,
}: SubscriptionsTableClientProps) {
	const [data, setData] = useState<Subscription[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);

	async function fetchData(pageNum: number) {
		setLoading(true);
		const res = await getAdminSubscriptions(pageNum);
		if (res.success && res.data) {
			setData(res.data.items);
			setTotal(res.data.total);
		}
		setLoading(false);
	}

	function handlePageChange(newPage: number) {
		setPage(newPage);
		if (newPage > 1) fetchData(newPage);
	}

	const columns: Column<Subscription>[] = [
		{
			key: "user",
			header: "Usuario",
			cell: (row) => (
				<div>
					<p className="font-medium">{row.userName}</p>
					<p className="text-xs text-muted-foreground">
						{row.userEmail}
					</p>
				</div>
			),
		},
		{
			key: "plan",
			header: "Plano",
			cell: (row) => <Badge variant="outline">{row.planName}</Badge>,
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
			key: "paymentMethod",
			header: "Pagamento",
			cell: (row) => (
				<span className="text-sm text-muted-foreground">
					{paymentLabels[row.paymentMethod] ?? row.paymentMethod}
				</span>
			),
		},
		{
			key: "currentPeriodEnd",
			header: "Vencimento",
			cell: (row) =>
				new Date(row.currentPeriodEnd).toLocaleDateString("pt-BR"),
		},
	];

	return (
		<DataTable
			columns={columns}
			data={data}
			loading={loading}
			emptyIcon={CreditCard}
			emptyTitle="Nenhuma assinatura"
			emptyDescription="Nenhuma assinatura encontrada."
			searchPlaceholder="Buscar assinaturas..."
			page={page}
			pageSize={20}
			totalCount={total}
			onPageChange={handlePageChange}
		/>
	);
}
