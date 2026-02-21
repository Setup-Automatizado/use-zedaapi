"use client";

import { useState, useRef } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Clock, CheckCircle, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
	getWaitlist,
	approveWaitlist,
	rejectWaitlist,
} from "@/server/actions/admin";

interface WaitlistEntry {
	id: string;
	email: string;
	name: string | null;
	status: string;
	createdAt: Date;
}

const statusConfig: Record<string, { label: string; className: string }> = {
	pending: { label: "Pendente", className: "bg-chart-2/10 text-chart-2" },
	approved: { label: "Aprovado", className: "bg-primary/10 text-primary" },
	rejected: {
		label: "Rejeitado",
		className: "bg-destructive/10 text-destructive",
	},
};

interface WaitlistTableClientProps {
	initialData: WaitlistEntry[];
	initialTotal: number;
}

export function WaitlistTableClient({
	initialData,
	initialTotal,
}: WaitlistTableClientProps) {
	const [data, setData] = useState<WaitlistEntry[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);
	const [search, setSearch] = useState("");
	const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);
	const [approveDialog, setApproveDialog] = useState<{
		open: boolean;
		entryId: string | null;
	}>({ open: false, entryId: null });
	const [rejectDialog, setRejectDialog] = useState<{
		open: boolean;
		entryId: string | null;
	}>({ open: false, entryId: null });
	const [actionLoading, setActionLoading] = useState(false);

	async function fetchData(pageNum: number, searchTerm?: string) {
		setLoading(true);
		const res = await getWaitlist(pageNum, searchTerm || undefined);
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

	const handleApprove = async () => {
		if (!approveDialog.entryId) return;
		setActionLoading(true);
		const res = await approveWaitlist(approveDialog.entryId);
		setActionLoading(false);
		if (res.success) {
			toast.success("Solicitação aprovada");
			fetchData(page, search);
		} else {
			toast.error("Erro ao aprovar");
		}
		setApproveDialog({ open: false, entryId: null });
	};

	const handleReject = async () => {
		if (!rejectDialog.entryId) return;
		setActionLoading(true);
		const res = await rejectWaitlist(rejectDialog.entryId);
		setActionLoading(false);
		if (res.success) {
			toast.success("Solicitação rejeitada");
			fetchData(page, search);
		} else {
			toast.error("Erro ao rejeitar");
		}
		setRejectDialog({ open: false, entryId: null });
	};

	const columns: Column<WaitlistEntry>[] = [
		{
			key: "email",
			header: "Email",
			cell: (row) => (
				<div>
					<p className="font-medium">{row.email}</p>
					{row.name && (
						<p className="text-xs text-muted-foreground">
							{row.name}
						</p>
					)}
				</div>
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
			header: "Data",
			cell: (row) => new Date(row.createdAt).toLocaleDateString("pt-BR"),
		},
		{
			key: "actions",
			header: "",
			cell: (row) =>
				row.status === "pending" ? (
					<div className="flex gap-1">
						<Button
							variant="ghost"
							size="icon-sm"
							className="text-primary"
							onClick={() =>
								setApproveDialog({
									open: true,
									entryId: row.id,
								})
							}
						>
							<CheckCircle className="size-4" />
						</Button>
						<Button
							variant="ghost"
							size="icon-sm"
							className="text-destructive"
							onClick={() =>
								setRejectDialog({
									open: true,
									entryId: row.id,
								})
							}
						>
							<XCircle className="size-4" />
						</Button>
					</div>
				) : null,
			className: "w-24",
		},
	];

	return (
		<>
			<DataTable
				columns={columns}
				data={data}
				loading={loading}
				emptyIcon={Clock}
				emptyTitle="Waitlist vazia"
				emptyDescription="Nenhuma solicitação pendente."
				onSearch={handleSearch}
				searchPlaceholder="Buscar na waitlist..."
				page={page}
				pageSize={20}
				totalCount={total}
				onPageChange={handlePageChange}
			/>

			<ConfirmDialog
				open={approveDialog.open}
				onOpenChange={(open) =>
					setApproveDialog({ open, entryId: null })
				}
				title="Aprovar solicitação"
				description="O usuário receberá um e-mail com o código de convite para criar sua conta."
				confirmLabel="Aprovar"
				loading={actionLoading}
				onConfirm={handleApprove}
			/>

			<ConfirmDialog
				open={rejectDialog.open}
				onOpenChange={(open) =>
					setRejectDialog({ open, entryId: null })
				}
				title="Rejeitar solicitação"
				description="O usuário será notificado que sua solicitação foi rejeitada."
				confirmLabel="Rejeitar"
				destructive
				loading={actionLoading}
				onConfirm={handleReject}
			/>
		</>
	);
}
