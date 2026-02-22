"use client";

import { useState, useCallback, useTransition, useMemo } from "react";
import Link from "next/link";
import { toast } from "sonner";
import {
	Plus,
	FileText,
	MoreHorizontal,
	Pencil,
	Globe,
	Archive,
	Trash2,
	Eye,
	FolderOpen,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
import {
	getAdminBlogPosts,
	publishBlogPost,
	archiveBlogPost,
	deleteBlogPost,
} from "@/server/actions/blog";

interface BlogPost {
	id: string;
	title: string;
	slug: string;
	status: string;
	category?: { id: string; name: string } | null;
	author?: { name: string | null } | null;
	publishedAt?: string | null;
	createdAt: string;
	viewCount?: number;
	tags?: Array<{ tag: { id: string; name: string } }>;
}

interface BlogTableClientProps {
	initialItems: BlogPost[];
	initialTotal: number;
}

function StatusBadge({ status }: { status: string }) {
	switch (status) {
		case "published":
			return (
				<Badge className="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20">
					Publicado
				</Badge>
			);
		case "archived":
			return <Badge variant="secondary">Arquivado</Badge>;
		default:
			return <Badge variant="outline">Rascunho</Badge>;
	}
}

function formatDate(dateStr: string | null | undefined): string {
	if (!dateStr) return "-";
	return new Date(dateStr).toLocaleDateString("pt-BR", {
		day: "2-digit",
		month: "short",
		year: "numeric",
	});
}

export function BlogTableClient({
	initialItems,
	initialTotal,
}: BlogTableClientProps) {
	const [items, setItems] = useState<BlogPost[]>(initialItems);
	const [total, setTotal] = useState(initialTotal);
	const [page, setPage] = useState(1);
	const [search, setSearch] = useState("");
	const [loading, setLoading] = useState(false);
	const [isPending, startTransition] = useTransition();

	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [deleteTarget, setDeleteTarget] = useState<BlogPost | null>(null);
	const [deleteLoading, setDeleteLoading] = useState(false);

	const fetchPosts = useCallback(async (p: number, q: string) => {
		setLoading(true);
		const res = await getAdminBlogPosts(p, q || undefined);
		if (res.success && res.data) {
			setItems(res.data.items as unknown as BlogPost[]);
			setTotal(res.data.total);
		}
		setLoading(false);
	}, []);

	const handlePageChange = useCallback(
		(newPage: number) => {
			setPage(newPage);
			fetchPosts(newPage, search);
		},
		[fetchPosts, search],
	);

	const handleSearch = useCallback(
		(query: string) => {
			setSearch(query);
			setPage(1);
			fetchPosts(1, query);
		},
		[fetchPosts],
	);

	const handlePublish = useCallback(
		(post: BlogPost) => {
			startTransition(async () => {
				const res = await publishBlogPost(post.id);
				if (res.success) {
					toast.success("Post publicado com sucesso");
					fetchPosts(page, search);
				} else {
					toast.error(res.error ?? "Erro ao publicar post");
				}
			});
		},
		[fetchPosts, page, search],
	);

	const handleArchive = useCallback(
		(post: BlogPost) => {
			startTransition(async () => {
				const res = await archiveBlogPost(post.id);
				if (res.success) {
					toast.success("Post arquivado com sucesso");
					fetchPosts(page, search);
				} else {
					toast.error(res.error ?? "Erro ao arquivar post");
				}
			});
		},
		[fetchPosts, page, search],
	);

	const handleDeleteConfirm = useCallback(async () => {
		if (!deleteTarget) return;
		setDeleteLoading(true);
		const res = await deleteBlogPost(deleteTarget.id);
		if (res.success) {
			toast.success("Post excluido com sucesso");
			setDeleteDialogOpen(false);
			setDeleteTarget(null);
			fetchPosts(page, search);
		} else {
			toast.error(res.error ?? "Erro ao excluir post");
		}
		setDeleteLoading(false);
	}, [deleteTarget, fetchPosts, page, search]);

	const columns = useMemo<Column<BlogPost>[]>(
		() => [
			{
				key: "title",
				header: "Titulo",
				cell: (row) => (
					<div className="min-w-0">
						<div className="flex items-center gap-2">
							<span className="truncate font-medium">
								{row.title}
							</span>
							<StatusBadge status={row.status} />
						</div>
						<p className="mt-0.5 truncate text-xs text-muted-foreground">
							/blog/{row.slug}
						</p>
					</div>
				),
				className: "max-w-[400px]",
			},
			{
				key: "category",
				header: "Categoria",
				cell: (row) => (
					<span className="text-sm text-muted-foreground">
						{row.category?.name ?? "-"}
					</span>
				),
			},
			{
				key: "author",
				header: "Autor",
				cell: (row) => (
					<span className="text-sm text-muted-foreground">
						{row.author?.name ?? "-"}
					</span>
				),
			},
			{
				key: "viewCount",
				header: "Views",
				cell: (row) => (
					<div className="flex items-center gap-1 text-sm text-muted-foreground">
						<Eye className="size-3.5" />
						{row.viewCount ?? 0}
					</div>
				),
				sortable: true,
			},
			{
				key: "publishedAt",
				header: "Publicado em",
				cell: (row) => (
					<span className="text-sm text-muted-foreground">
						{formatDate(row.publishedAt)}
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
							<DropdownMenuItem asChild>
								<Link href={`/admin/blog/${row.id}/editar`}>
									<Pencil className="mr-2 size-4" />
									Editar
								</Link>
							</DropdownMenuItem>
							{row.status !== "published" && (
								<DropdownMenuItem
									onClick={() => handlePublish(row)}
									disabled={isPending}
								>
									<Globe className="mr-2 size-4" />
									Publicar
								</DropdownMenuItem>
							)}
							{row.status === "published" && (
								<DropdownMenuItem
									onClick={() => handleArchive(row)}
									disabled={isPending}
								>
									<Archive className="mr-2 size-4" />
									Arquivar
								</DropdownMenuItem>
							)}
							{row.status === "published" && (
								<DropdownMenuItem asChild>
									<Link
										href={`/blog/${row.slug}`}
										target="_blank"
									>
										<Eye className="mr-2 size-4" />
										Ver no site
									</Link>
								</DropdownMenuItem>
							)}
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
		[handlePublish, handleArchive, isPending],
	);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Blog"
				description="Gerencie os posts do blog."
				action={
					<div className="flex gap-2">
						<Button variant="outline" asChild>
							<Link href="/admin/blog/categorias">
								<FolderOpen className="size-4" />
								Categorias
							</Link>
						</Button>
						<Button asChild>
							<Link href="/admin/blog/novo">
								<Plus className="size-4" />
								Novo Post
							</Link>
						</Button>
					</div>
				}
			/>

			<DataTable
				columns={columns}
				data={items}
				loading={loading}
				searchPlaceholder="Buscar posts..."
				onSearch={handleSearch}
				emptyIcon={FileText}
				emptyTitle="Nenhum post encontrado"
				emptyDescription="Crie seu primeiro post clicando no botao acima."
				page={page}
				pageSize={20}
				totalCount={total}
				onPageChange={handlePageChange}
			/>

			<ConfirmDialog
				open={deleteDialogOpen}
				onOpenChange={setDeleteDialogOpen}
				title="Excluir post"
				description={`Tem certeza que deseja excluir "${deleteTarget?.title}"? Esta acao nao pode ser desfeita.`}
				confirmLabel="Excluir"
				destructive
				loading={deleteLoading}
				onConfirm={handleDeleteConfirm}
			/>
		</div>
	);
}
