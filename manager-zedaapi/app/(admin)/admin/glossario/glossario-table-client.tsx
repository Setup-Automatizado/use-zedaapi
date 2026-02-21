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
import { MoreHorizontal, Plus, Pencil, Trash2, BookA } from "lucide-react";
import { toast } from "sonner";
import {
	getAdminGlossaryTerms,
	deleteGlossaryTerm,
} from "@/server/actions/content";

interface GlossaryTerm {
	id: string;
	term: string;
	slug: string;
	definition: string;
	status: string;
	createdAt: Date;
}

interface GlossarioTableClientProps {
	initialData: GlossaryTerm[];
	initialTotal: number;
}

export function GlossarioTableClient({
	initialData,
	initialTotal,
}: GlossarioTableClientProps) {
	const [terms, setTerms] = useState<GlossaryTerm[]>(initialData);
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
		const res = await getAdminGlossaryTerms(
			pageNum,
			searchTerm || undefined,
		);
		if (res.success && res.data) {
			setTerms(res.data.items as unknown as GlossaryTerm[]);
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
		const res = await deleteGlossaryTerm(deleteDialog.id);
		setDeleteLoading(false);
		if (res.success) {
			toast.success("Termo excluido com sucesso");
			fetchData(page, search);
		} else {
			toast.error(res.error ?? "Erro ao excluir termo");
		}
		setDeleteDialog({ open: false, id: null });
	}

	const columns: Column<GlossaryTerm>[] = [
		{
			key: "term",
			header: "Termo",
			cell: (row) => (
				<div>
					<p className="font-medium">{row.term}</p>
					<p className="text-xs text-muted-foreground">
						/glossario/{row.slug}
					</p>
				</div>
			),
		},
		{
			key: "definition",
			header: "Definicao",
			cell: (row) => (
				<p className="line-clamp-1 max-w-xs text-sm text-muted-foreground">
					{row.definition}
				</p>
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
							<Link href={`/admin/glossario/${row.id}/editar`}>
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
				data={terms}
				loading={loading}
				emptyIcon={BookA}
				emptyTitle="Nenhum termo encontrado"
				emptyDescription="Crie seu primeiro termo clicando no botao acima."
				emptyActionLabel="Novo Termo"
				onEmptyAction={() => {
					window.location.href = "/admin/glossario/novo";
				}}
				onSearch={handleSearch}
				searchPlaceholder="Buscar termos..."
				page={page}
				pageSize={20}
				totalCount={total}
				onPageChange={handlePageChange}
				headerAction={
					<Button asChild>
						<Link href="/admin/glossario/novo">
							<Plus className="size-4" />
							Novo Termo
						</Link>
					</Button>
				}
			/>

			<ConfirmDialog
				open={deleteDialog.open}
				onOpenChange={(open) => setDeleteDialog({ open, id: null })}
				title="Excluir termo"
				description="Tem certeza que deseja excluir este termo? Esta acao nao pode ser desfeita."
				confirmLabel="Excluir"
				destructive
				loading={deleteLoading}
				onConfirm={handleDeleteConfirm}
			/>
		</>
	);
}
