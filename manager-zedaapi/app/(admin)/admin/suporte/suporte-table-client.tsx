"use client";

import { useState, useRef } from "react";
import Link from "next/link";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { MoreHorizontal, Plus, Pencil, Trash2, HelpCircle } from "lucide-react";
import { toast } from "sonner";
import {
	getAdminSupportArticles,
	deleteSupportArticle,
} from "@/server/actions/content";

interface SupportArticle {
	id: string;
	title: string;
	slug: string;
	status: string;
	sortOrder: number;
	viewCount: number;
	category: { id: string; name: string } | null;
	createdAt: Date;
}

interface SuporteTableClientProps {
	initialData: SupportArticle[];
	initialTotal: number;
}

export function SuporteTableClient({
	initialData,
	initialTotal,
}: SuporteTableClientProps) {
	const [articles, setArticles] = useState<SupportArticle[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);
	const [search, setSearch] = useState("");
	const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);
	const [deleteDialog, setDeleteDialog] = useState<{
		open: boolean;
		id: string | null;
	}>({ open: false, id: null });
	const [deleteLoading, setDeleteLoading] = useState(false);

	async function fetchData(pageNum: number, searchTerm?: string) {
		setLoading(true);
		const res = await getAdminSupportArticles(
			pageNum,
			searchTerm || undefined,
		);
		if (res.success && res.data) {
			setArticles(res.data.items as unknown as SupportArticle[]);
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

	async function handleDeleteConfirm() {
		if (!deleteDialog.id) return;
		setDeleteLoading(true);
		const res = await deleteSupportArticle(deleteDialog.id);
		setDeleteLoading(false);
		if (res.success) {
			toast.success("Artigo excluido com sucesso");
			fetchData(page, search);
		} else {
			toast.error(res.error ?? "Erro ao excluir artigo");
		}
		setDeleteDialog({ open: false, id: null });
	}

	const columns: Column<SupportArticle>[] = [
		{
			key: "title",
			header: "Titulo",
			cell: (row) => (
				<div>
					<p className="font-medium">{row.title}</p>
					<p className="text-xs text-muted-foreground">
						/suporte/{row.slug}
					</p>
				</div>
			),
		},
		{
			key: "category",
			header: "Categoria",
			cell: (row) => (
				<span className="text-sm text-muted-foreground">
					{row.category?.name ?? "Sem categoria"}
				</span>
			),
		},
		{
			key: "status",
			header: "Status",
			cell: (row) => (
				<Badge
					variant={row.status === "published" ? "default" : "outline"}
				>
					{row.status === "published" ? "Publicado" : "Rascunho"}
				</Badge>
			),
		},
		{
			key: "viewCount",
			header: "Views",
			cell: (row) => (
				<span className="text-sm tabular-nums text-muted-foreground">
					{(row.viewCount ?? 0).toLocaleString("pt-BR")}
				</span>
			),
			sortable: true,
		},
		{
			key: "sortOrder",
			header: "Ordem",
			cell: (row) => (
				<span className="text-sm tabular-nums text-muted-foreground">
					{row.sortOrder}
				</span>
			),
			sortable: true,
		},
		{
			key: "actions",
			header: "",
			cell: (row) => (
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button variant="ghost" size="icon-sm">
							<MoreHorizontal className="size-4" />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end">
						<DropdownMenuItem asChild>
							<Link href={`/admin/suporte/${row.id}/editar`}>
								<Pencil className="size-4" />
								Editar
							</Link>
						</DropdownMenuItem>
						<DropdownMenuSeparator />
						<DropdownMenuItem
							className="text-destructive"
							onClick={() =>
								setDeleteDialog({
									open: true,
									id: row.id,
								})
							}
						>
							<Trash2 className="size-4" />
							Excluir
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			),
			className: "w-12",
		},
	];

	return (
		<>
			<DataTable
				columns={columns}
				data={articles}
				loading={loading}
				emptyIcon={HelpCircle}
				emptyTitle="Nenhum artigo de suporte"
				emptyDescription="Crie seu primeiro artigo para a central de ajuda."
				emptyActionLabel="Novo Artigo"
				onEmptyAction={() => {
					window.location.href = "/admin/suporte/novo";
				}}
				onSearch={handleSearch}
				searchPlaceholder="Buscar artigos..."
				page={page}
				pageSize={20}
				totalCount={total}
				onPageChange={handlePageChange}
				headerAction={
					<Button asChild>
						<Link href="/admin/suporte/novo">
							<Plus className="size-4" />
							Novo Artigo
						</Link>
					</Button>
				}
			/>

			<ConfirmDialog
				open={deleteDialog.open}
				onOpenChange={(open) => setDeleteDialog({ open, id: null })}
				title="Excluir artigo"
				description="Tem certeza que deseja excluir este artigo? Esta acao nao pode ser desfeita."
				confirmLabel="Excluir"
				destructive
				loading={deleteLoading}
				onConfirm={handleDeleteConfirm}
			/>
		</>
	);
}
