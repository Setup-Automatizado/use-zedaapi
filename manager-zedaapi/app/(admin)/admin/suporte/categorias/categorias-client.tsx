"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogFooter,
} from "@/components/ui/dialog";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import {
	Plus,
	MoreHorizontal,
	Pencil,
	Trash2,
	FolderOpen,
	Loader2,
} from "lucide-react";
import { slugify } from "@/lib/slugify";
import {
	createSupportCategory,
	updateSupportCategory,
	deleteSupportCategory,
} from "@/server/actions/content";

const ICON_OPTIONS = [
	{ value: "help-circle", label: "Ajuda" },
	{ value: "book-open", label: "Documentacao" },
	{ value: "settings", label: "Configuracoes" },
	{ value: "credit-card", label: "Pagamentos" },
	{ value: "shield", label: "Seguranca" },
	{ value: "zap", label: "Integracao" },
	{ value: "users", label: "Usuarios" },
	{ value: "message-square", label: "Mensagens" },
	{ value: "code", label: "API" },
	{ value: "smartphone", label: "WhatsApp" },
];

interface Category {
	id: string;
	name: string;
	slug: string;
	description: string | null;
	icon: string | null;
	sortOrder: number;
	_count: { articles: number };
}

interface CategoriasClientProps {
	initialCategories: Category[];
}

interface CategoryFormState {
	name: string;
	slug: string;
	description: string;
	icon: string;
	sortOrder: string;
}

const emptyForm: CategoryFormState = {
	name: "",
	slug: "",
	description: "",
	icon: "",
	sortOrder: "0",
};

export function CategoriasClient({ initialCategories }: CategoriasClientProps) {
	const router = useRouter();
	const [categories, setCategories] = useState<Category[]>(initialCategories);
	const [formOpen, setFormOpen] = useState(false);
	const [editingId, setEditingId] = useState<string | null>(null);
	const [form, setForm] = useState<CategoryFormState>(emptyForm);
	const [saving, setSaving] = useState(false);
	const [deleteDialog, setDeleteDialog] = useState<{
		open: boolean;
		id: string | null;
		articleCount: number;
	}>({ open: false, id: null, articleCount: 0 });
	const [deleteLoading, setDeleteLoading] = useState(false);

	function openCreate() {
		setEditingId(null);
		setForm(emptyForm);
		setFormOpen(true);
	}

	function openEdit(cat: Category) {
		setEditingId(cat.id);
		setForm({
			name: cat.name,
			slug: cat.slug,
			description: cat.description ?? "",
			icon: cat.icon ?? "",
			sortOrder: String(cat.sortOrder),
		});
		setFormOpen(true);
	}

	function handleNameChange(name: string) {
		setForm((prev) => ({
			...prev,
			name,
			slug: editingId ? prev.slug : slugify(name),
		}));
	}

	async function handleSubmit() {
		if (!form.name.trim()) {
			toast.error("O nome e obrigatorio");
			return;
		}

		setSaving(true);

		if (editingId) {
			const res = await updateSupportCategory(editingId, {
				name: form.name.trim(),
				slug: form.slug || slugify(form.name),
				description: form.description.trim() || undefined,
				icon: form.icon || undefined,
				sortOrder: Number.parseInt(form.sortOrder, 10) || 0,
			});
			setSaving(false);
			if (res.success) {
				toast.success("Categoria atualizada com sucesso");
				setFormOpen(false);
				router.refresh();
			} else {
				toast.error(res.error ?? "Erro ao atualizar categoria");
			}
		} else {
			const res = await createSupportCategory({
				name: form.name.trim(),
				slug: form.slug || slugify(form.name),
				description: form.description.trim() || undefined,
				icon: form.icon || undefined,
				sortOrder: Number.parseInt(form.sortOrder, 10) || 0,
			});
			setSaving(false);
			if (res.success) {
				toast.success("Categoria criada com sucesso");
				setFormOpen(false);
				router.refresh();
			} else {
				toast.error(res.error ?? "Erro ao criar categoria");
			}
		}
	}

	async function handleDeleteConfirm() {
		if (!deleteDialog.id) return;
		setDeleteLoading(true);
		const res = await deleteSupportCategory(deleteDialog.id);
		setDeleteLoading(false);
		if (res.success) {
			toast.success("Categoria excluida com sucesso");
			setCategories((prev) =>
				prev.filter((c) => c.id !== deleteDialog.id),
			);
			router.refresh();
		} else {
			toast.error(res.error ?? "Erro ao excluir categoria");
		}
		setDeleteDialog({ open: false, id: null, articleCount: 0 });
	}

	const columns: Column<Category>[] = [
		{
			key: "name",
			header: "Nome",
			cell: (row) => (
				<div>
					<p className="font-medium">{row.name}</p>
					<p className="text-xs text-muted-foreground">/{row.slug}</p>
				</div>
			),
		},
		{
			key: "description",
			header: "Descricao",
			cell: (row) => (
				<span className="line-clamp-1 text-sm text-muted-foreground">
					{row.description || "-"}
				</span>
			),
		},
		{
			key: "icon",
			header: "Icone",
			cell: (row) => (
				<span className="text-sm text-muted-foreground">
					{row.icon || "-"}
				</span>
			),
		},
		{
			key: "articles",
			header: "Artigos",
			cell: (row) => (
				<span className="text-sm tabular-nums text-muted-foreground">
					{row._count.articles}
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
						<DropdownMenuItem onClick={() => openEdit(row)}>
							<Pencil className="size-4" />
							Editar
						</DropdownMenuItem>
						<DropdownMenuSeparator />
						<DropdownMenuItem
							className="text-destructive"
							onClick={() =>
								setDeleteDialog({
									open: true,
									id: row.id,
									articleCount: row._count.articles,
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
				data={categories}
				emptyIcon={FolderOpen}
				emptyTitle="Nenhuma categoria"
				emptyDescription="Crie sua primeira categoria para organizar os artigos de suporte."
				emptyActionLabel="Nova Categoria"
				onEmptyAction={openCreate}
				headerAction={
					<Button onClick={openCreate}>
						<Plus className="size-4" />
						Nova Categoria
					</Button>
				}
			/>

			<Dialog open={formOpen} onOpenChange={setFormOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{editingId ? "Editar Categoria" : "Nova Categoria"}
						</DialogTitle>
					</DialogHeader>

					<div className="space-y-4 py-2">
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
							<Label>Slug</Label>
							<Input
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
							<Label htmlFor="cat-description">Descricao</Label>
							<Textarea
								id="cat-description"
								value={form.description}
								onChange={(e) =>
									setForm((prev) => ({
										...prev,
										description: e.target.value,
									}))
								}
								placeholder="Breve descricao da categoria..."
								rows={2}
								className="min-h-0"
							/>
						</div>

						<div className="grid gap-4 sm:grid-cols-2">
							<div className="space-y-1.5">
								<Label>Icone</Label>
								<Select
									value={form.icon}
									onValueChange={(value) =>
										setForm((prev) => ({
											...prev,
											icon: value,
										}))
									}
								>
									<SelectTrigger>
										<SelectValue placeholder="Selecione um icone" />
									</SelectTrigger>
									<SelectContent>
										{ICON_OPTIONS.map((opt) => (
											<SelectItem
												key={opt.value}
												value={opt.value}
											>
												{opt.label}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>

							<div className="space-y-1.5">
								<Label htmlFor="cat-sortOrder">Ordem</Label>
								<Input
									id="cat-sortOrder"
									type="number"
									value={form.sortOrder}
									onChange={(e) =>
										setForm((prev) => ({
											...prev,
											sortOrder: e.target.value,
										}))
									}
									placeholder="0"
								/>
							</div>
						</div>
					</div>

					<DialogFooter>
						<Button
							type="button"
							variant="outline"
							onClick={() => setFormOpen(false)}
							disabled={saving}
						>
							Cancelar
						</Button>
						<Button
							type="button"
							onClick={handleSubmit}
							disabled={saving}
						>
							{saving && (
								<Loader2 className="size-4 animate-spin" />
							)}
							{editingId ? "Salvar" : "Criar"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<ConfirmDialog
				open={deleteDialog.open}
				onOpenChange={(open) =>
					setDeleteDialog({ open, id: null, articleCount: 0 })
				}
				title="Excluir categoria"
				description={
					deleteDialog.articleCount > 0
						? `Esta categoria possui ${deleteDialog.articleCount} artigo(s). Mova ou exclua os artigos antes de excluir a categoria.`
						: "Tem certeza que deseja excluir esta categoria? Esta acao nao pode ser desfeita."
				}
				confirmLabel="Excluir"
				destructive
				loading={deleteLoading}
				onConfirm={handleDeleteConfirm}
			/>
		</>
	);
}
