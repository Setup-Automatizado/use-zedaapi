"use client";

import { useState, useCallback, useEffect } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Receipt } from "lucide-react";
import { cn } from "@/lib/utils";
import { getAdminInvoices } from "@/server/actions/admin";

interface Invoice {
	id: string;
	userName: string;
	amount: number;
	status: string;
	paymentMethod: string | null;
	nfseStatus: string | null;
	createdAt: Date;
}

const statusConfig: Record<string, { label: string; className: string }> = {
	paid: { label: "Pago", className: "bg-primary/10 text-primary" },
	pending: { label: "Pendente", className: "bg-chart-2/10 text-chart-2" },
	draft: { label: "Rascunho", className: "bg-muted text-muted-foreground" },
	overdue: {
		label: "Vencido",
		className: "bg-destructive/10 text-destructive",
	},
	refunded: {
		label: "Reembolsado",
		className: "bg-muted text-muted-foreground",
	},
};

const nfseConfig: Record<string, { label: string; className: string }> = {
	issued: { label: "Emitida", className: "bg-primary/10 text-primary" },
	pending: { label: "Pendente", className: "bg-chart-2/10 text-chart-2" },
	processing: {
		label: "Processando",
		className: "bg-chart-2/10 text-chart-2",
	},
	error: { label: "Erro", className: "bg-destructive/10 text-destructive" },
	canceled: {
		label: "Cancelada",
		className: "bg-muted text-muted-foreground",
	},
};

const paymentLabels: Record<string, string> = {
	stripe: "Cartao",
	pix: "PIX",
	boleto: "Boleto",
};

interface InvoicesTableClientProps {
	initialData: Invoice[];
	initialTotal: number;
}

export function InvoicesTableClient({
	initialData,
	initialTotal,
}: InvoicesTableClientProps) {
	const [data, setData] = useState<Invoice[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);

	const load = useCallback(async () => {
		setLoading(true);
		const res = await getAdminInvoices(page);
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

	const columns: Column<Invoice>[] = [
		{
			key: "id",
			header: "Fatura",
			cell: (row) => (
				<span className="font-mono text-xs">
					{row.id.slice(0, 12)}...
				</span>
			),
		},
		{
			key: "user",
			header: "Usuario",
			cell: (row) => <span className="font-medium">{row.userName}</span>,
		},
		{
			key: "amount",
			header: "Valor",
			cell: (row) =>
				`R$ ${row.amount.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}`,
			className: "text-right",
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
					{row.paymentMethod
						? (paymentLabels[row.paymentMethod] ??
							row.paymentMethod)
						: "-"}
				</span>
			),
		},
		{
			key: "nfseStatus",
			header: "NFS-e",
			cell: (row) => {
				if (!row.nfseStatus) {
					return (
						<span className="text-xs text-muted-foreground">-</span>
					);
				}
				const config = nfseConfig[row.nfseStatus] ?? {
					label: row.nfseStatus,
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
			header: "Data",
			cell: (row) => new Date(row.createdAt).toLocaleDateString("pt-BR"),
		},
	];

	return (
		<DataTable
			columns={columns}
			data={data}
			loading={loading}
			emptyIcon={Receipt}
			emptyTitle="Nenhuma fatura"
			emptyDescription="Nenhuma fatura encontrada."
			searchPlaceholder="Buscar faturas..."
			page={page}
			pageSize={20}
			totalCount={total}
			onPageChange={setPage}
		/>
	);
}
