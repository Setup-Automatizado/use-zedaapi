"use client";

import { useState, useCallback, useMemo } from "react";
import { toast } from "sonner";
import { FolderOpen, Plus, Pencil, Trash2, MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Loader2 } from "lucide-react";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { DataTable, type Column } from "@/components/shared/data-table";
import { PageHeader } from "@/components/shared/page-header";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { slugify } from "@/lib/slugify";
import {
	getAdminBlogCategories,
	createBlogCategory,
	updateBlogCategory,
	deleteBlogCategory,
} from "@/server/actions/blog";

interface Category {
	id: string;
	name: string;
	slug: string;
	description: string | null;
	sortOrder: number;
	_count: { posts: number };
}

interface CategoriasClientProps {
	initialCategories: Category[];
}

interface CategoryFormState {
	name: string;
	slug: string;
	description: string;
	sortOrder: number;
}

const emptyForm: CategoryFormState = {
	name: "",
	slug: "",
	description: "",
	sortOrder: 0,
};

export function CategoriasClient({ initialCategories }: CategoriasClientProps) {
	const [categories, setCategories] = useState<Category[]>(initialCategories);
	const [formDialogOpen, setFormDialogOpen] = useState(false);
	const [editingCategory, setEditingCategory] = useState<Category | null>(
		null,
	);
	const [form, setForm] = useState<CategoryFormState>(emptyForm);
	const [saving, setSaving] = useState(false);

	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [deleteTarget, setDeleteTarget] = useState<Category | null>(null);
	const [deleteLoading, setDeleteLoading] = useState(false);

	const refreshCategories = useCallback(async () => {
		const res = await getAdminBlogCategories();
		if (res.success && res.data) {
			setCategories(res.data as unknown as Category[]);
		}
	}, []);

	const openCreateDialog = useCallback(() => {
		setEditingCategory(null);
		setForm(emptyForm);
		setFormDialogOpen(true);
	}, []);

	const openEditDialog = useCallback((cat: Category) => {
		setEditingCategory(cat);
		setForm({
			name: cat.name,
			slug: cat.slug,
			description: cat.description ?? "",
			sortOrder: cat.sortOrder,
		});
		setFormDialogOpen(true);
	}, []);

	const handleNameChange = useCallback(
		(value: string) => {
			setForm((prev) => ({
				...prev,
				name: value,
				// Auto-generate slug only when creating
				slug: editingCategory ? prev.slug : slugify(value),
			}));
		},
		[editingCategory],
	);

	const handleFormSubmit = useCallback(async () => {
		if (!form.name.trim()) {
			toast.error("O nome e obrigatorio");
			return;
		}

		setSaving(true);

		if (editingCategory) {
			const res = await updateBlogCategory(editingCategory.id, {
				name: form.name.trim(),
				slug: form.slug || slugify(form.name),
				description: form.description.trim() || undefined,
				sortOrder: form.sortOrder,
			});
			if (res.success) {
				toast.success("Categoria atualizada com sucesso");
				setFormDialogOpen(false);
				await refreshCategories();
			} else {
				toast.error(res.error ?? "Erro ao atualizar categoria");
			}
		} else {
			const res = await createBlogCategory({
				name: form.name.trim(),
				slug: form.slug || slugify(form.name),
				description: form.description.trim() || undefined,
				sortOrder: form.sortOrder,
			});
			if (res.success) {
				toast.success("Categoria criada com sucesso");
				setFormDialogOpen(false);
				await refreshCategories();
			} else {
				toast.error(res.error ?? "Erro ao criar categoria");
			}
		}

		setSaving(false);
	}, [form, editingCategory, refreshCategories]);

	const handleDeleteConfirm = useCallback(async () => {
		if (!deleteTarget) return;
		setDeleteLoading(true);

		const res = await deleteBlogCategory(deleteTarget.id);
		if (res.success) {
			toast.success("Categoria excluida com sucesso");
			setDeleteDialogOpen(false);
			setDeleteTarget(null);
			await refreshCategories();
		} else {
			toast.error(res.error ?? "Erro ao excluir categoria");
		}

		setDeleteLoading(false);
	}, [deleteTarget, refreshCategories]);

	const columns = useMemo<Column<Category>[]>(
		() => [
			{
				key: "name",
				header: "Nome",
				cell: (row) => (
					<div className="min-w-0">
						<span className="font-medium">{row.name}</span>
						<p className="mt-0.5 truncate text-xs text-muted-foreground">
							/{row.slug}
						</p>
					</div>
				),
				className: "max-w-[250px]",
			},
			{
				key: "description",
				header: "Descricao",
				cell: (row) => (
					<span className="line-clamp-1 text-sm text-muted-foreground">
						{row.description ?? "-"}
					</span>
				),
				className: "max-w-[300px]",
			},
			{
				key: "posts",
				header: "Posts",
				cell: (row) => (
					<span className="text-sm text-muted-foreground">
						{row._count.posts}
					</span>
				),
				sortable: true,
			},
			{
				key: "sortOrder",
				header: "Ordem",
				cell: (row) => (
					<span className="text-sm text-muted-foreground">
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
							<Button variant="ghost" size="icon-xs">
								<MoreHorizontal className="size-4" />
								<span className="sr-only">Acoes</span>
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								onClick={() => openEditDialog(row)}
							>
								<Pencil className="mr-2 size-4" />
								Editar
							</DropdownMenuItem>
							<DropdownMenuSeparator />
							<DropdownMenuItem
								className="text-destructive focus:text-destructive"
								onClick={() => {
									setDeleteTarget(row);
									setDeleteDialogOpen(true);
								}}
							>
								<Trash2 className="mr-2 size-4" />
								Excluir
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				),
				className: "w-[50px]",
			},
		],
		[openEditDialog],
	);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Categorias do Blog"
				description="Gerencie as categorias dos posts."
				backHref="/admin/blog"
				action={
					<Button onClick={openCreateDialog}>
						<Plus className="size-4" />
						Nova Categoria
					</Button>
				}
			/>

			<DataTable
				columns={columns}
				data={categories}
				emptyIcon={FolderOpen}
				emptyTitle="Nenhuma categoria encontrada"
				emptyDescription="Crie sua primeira categoria clicando no botao acima."
			/>

			{/* Form Dialog */}
			<Dialog open={formDialogOpen} onOpenChange={setFormDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{editingCategory
								? "Editar Categoria"
								: "Nova Categoria"}
						</DialogTitle>
					</DialogHeader>

					<div className="space-y-4">
						<div className="space-y-1.5">
							<Label htmlFor="cat-name">Nome</Label>
							<Input
								id="cat-name"
								value={form.name}
								onChange={(e) =>
									handleNameChange(e.target.value)
								}
								placeholder="Nome da categoria"
							/>
						</div>

						<div className="space-y-1.5">
							<Label htmlFor="cat-slug">Slug</Label>
							<Input
								id="cat-slug"
								value={form.slug}
								onChange={(e) =>
									setForm((prev) => ({
										...prev,
										slug: slugify(e.target.value),
									}))
								}
								placeholder="slug-da-categoria"
								className="font-mono text-sm"
							/>
						</div>

						<div className="space-y-1.5">
							<Label htmlFor="cat-desc">Descricao</Label>
							<Textarea
								id="cat-desc"
								value={form.description}
								onChange={(e) =>
									setForm((prev) => ({
										...prev,
										description: e.target.value,
									}))
								}
								placeholder="Descricao opcional da categoria..."
								rows={3}
								className="min-h-0"
							/>
						</div>

						<div className="space-y-1.5">
							<Label htmlFor="cat-order">Ordem de exibicao</Label>
							<Input
								id="cat-order"
								type="number"
								value={form.sortOrder}
								onChange={(e) =>
									setForm((prev) => ({
										...prev,
										sortOrder:
											Number.parseInt(
												e.target.value,
												10,
											) || 0,
									}))
								}
								min={0}
							/>
						</div>
					</div>

					<DialogFooter className="mt-4">
						<Button
							variant="outline"
							onClick={() => setFormDialogOpen(false)}
							disabled={saving}
						>
							Cancelar
						</Button>
						<Button
							onClick={() => void handleFormSubmit()}
							disabled={saving}
						>
							{saving && (
								<Loader2 className="size-4 animate-spin" />
							)}
							{editingCategory ? "Salvar" : "Criar"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Delete Confirm */}
			<ConfirmDialog
				open={deleteDialogOpen}
				onOpenChange={setDeleteDialogOpen}
				title="Excluir categoria"
				description={`Tem certeza que deseja excluir "${deleteTarget?.name}"? ${
					deleteTarget && deleteTarget._count.posts > 0
						? `Esta categoria possui ${deleteTarget._count.posts} post(s) associado(s).`
						: "Esta acao nao pode ser desfeita."
				}`}
				confirmLabel="Excluir"
				destructive
				loading={deleteLoading}
				onConfirm={handleDeleteConfirm}
			/>
		</div>
	);
}
