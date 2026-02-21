"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { ContentEditor } from "@/components/admin/content-editor";
import { SlugInput } from "@/components/admin/slug-input";
import { SeoFields } from "@/components/admin/seo-fields";
import { MediaUpload } from "@/components/admin/media-upload";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Loader2 } from "lucide-react";
import {
	createSupportArticle,
	updateSupportArticle,
} from "@/server/actions/content";

interface SuporteFormClientProps {
	categories: Array<{ id: string; name: string }>;
	initialData?: {
		id: string;
		title: string;
		slug: string;
		content: string;
		excerpt: string | null;
		categoryId: string | null;
		seoTitle: string | null;
		seoDescription: string | null;
		status: string;
		sortOrder: number;
	};
}

export function SuporteFormClient({
	categories,
	initialData,
}: SuporteFormClientProps) {
	const router = useRouter();
	const isEditing = !!initialData;

	const [title, setTitle] = useState(initialData?.title ?? "");
	const [slug, setSlug] = useState(initialData?.slug ?? "");
	const [categoryId, setCategoryId] = useState(initialData?.categoryId ?? "");
	const [status, setStatus] = useState(initialData?.status ?? "draft");
	const [sortOrder, setSortOrder] = useState(
		String(initialData?.sortOrder ?? 0),
	);
	const [excerpt, setExcerpt] = useState(initialData?.excerpt ?? "");
	const [content, setContent] = useState(initialData?.content ?? "");
	const [seoTitle, setSeoTitle] = useState(initialData?.seoTitle ?? "");
	const [seoDescription, setSeoDescription] = useState(
		initialData?.seoDescription ?? "",
	);
	const [seoKeywords, setSeoKeywords] = useState("");
	const [mediaOpen, setMediaOpen] = useState(false);
	const [saving, setSaving] = useState(false);

	function handleSeoChange(
		field: "seoTitle" | "seoDescription" | "seoKeywords",
		value: string,
	) {
		if (field === "seoTitle") setSeoTitle(value);
		else if (field === "seoDescription") setSeoDescription(value);
		else setSeoKeywords(value);
	}

	function handleMediaInsert(media: {
		url: string;
		type: string;
		filename: string;
	}) {
		const isImage = media.type.startsWith("image/");
		const markdown = isImage
			? `![${media.filename}](${media.url})`
			: `[${media.filename}](${media.url})`;
		setContent((prev) => (prev ? `${prev}\n\n${markdown}` : markdown));
		setMediaOpen(false);
	}

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();

		if (!title.trim()) {
			toast.error("O titulo e obrigatorio");
			return;
		}
		if (!content.trim()) {
			toast.error("O conteudo e obrigatorio");
			return;
		}
		if (!categoryId) {
			toast.error("Selecione uma categoria");
			return;
		}

		setSaving(true);

		const formData = {
			title: title.trim(),
			content,
			excerpt: excerpt.trim() || undefined,
			categoryId,
			seoTitle: seoTitle.trim() || undefined,
			seoDescription: seoDescription.trim() || undefined,
			status,
			sortOrder: Number.parseInt(sortOrder, 10) || 0,
		};

		const res = isEditing
			? await updateSupportArticle(initialData.id, formData)
			: await createSupportArticle(formData);

		setSaving(false);

		if (res.success) {
			toast.success(
				isEditing
					? "Artigo atualizado com sucesso"
					: "Artigo criado com sucesso",
			);
			router.push("/admin/suporte");
		} else {
			toast.error(res.error ?? "Erro ao salvar artigo");
		}
	}

	return (
		<>
			<form onSubmit={handleSubmit} className="space-y-6">
				<Card>
					<CardHeader>
						<CardTitle>Informacoes basicas</CardTitle>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="space-y-1.5">
							<Label htmlFor="title">Titulo</Label>
							<Input
								id="title"
								value={title}
								onChange={(e) => setTitle(e.target.value)}
								placeholder="Titulo do artigo"
							/>
						</div>

						<SlugInput
							value={slug}
							onChange={setSlug}
							sourceValue={title}
							prefix="/suporte"
						/>

						<div className="grid gap-4 sm:grid-cols-2">
							<div className="space-y-1.5">
								<Label>Categoria</Label>
								<Select
									value={categoryId}
									onValueChange={setCategoryId}
								>
									<SelectTrigger>
										<SelectValue placeholder="Selecione uma categoria" />
									</SelectTrigger>
									<SelectContent>
										{categories.map((cat) => (
											<SelectItem
												key={cat.id}
												value={cat.id}
											>
												{cat.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>

							<div className="space-y-1.5">
								<Label>Status</Label>
								<Select
									value={status}
									onValueChange={setStatus}
								>
									<SelectTrigger>
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="draft">
											Rascunho
										</SelectItem>
										<SelectItem value="published">
											Publicado
										</SelectItem>
									</SelectContent>
								</Select>
							</div>
						</div>

						<div className="space-y-1.5">
							<Label htmlFor="sortOrder">Ordem</Label>
							<Input
								id="sortOrder"
								type="number"
								value={sortOrder}
								onChange={(e) => setSortOrder(e.target.value)}
								placeholder="0"
								className="max-w-[120px]"
							/>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>Conteudo</CardTitle>
					</CardHeader>
					<CardContent className="space-y-4">
						<div className="space-y-1.5">
							<Label htmlFor="excerpt">Resumo</Label>
							<Textarea
								id="excerpt"
								value={excerpt}
								onChange={(e) => setExcerpt(e.target.value)}
								placeholder="Breve descricao do artigo..."
								rows={3}
								className="min-h-0"
							/>
						</div>

						<div className="space-y-1.5">
							<Label>Conteudo</Label>
							<ContentEditor
								value={content}
								onChange={setContent}
								onInsertMedia={() => setMediaOpen(true)}
							/>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>SEO</CardTitle>
					</CardHeader>
					<CardContent>
						<SeoFields
							seoTitle={seoTitle}
							seoDescription={seoDescription}
							seoKeywords={seoKeywords}
							onChange={handleSeoChange}
							previewUrl={`https://zedaapi.com/suporte/${slug || "..."}`}
						/>
					</CardContent>
				</Card>

				<div className="flex justify-end gap-3">
					<Button
						type="button"
						variant="outline"
						onClick={() => router.push("/admin/suporte")}
						disabled={saving}
					>
						Cancelar
					</Button>
					<Button type="submit" disabled={saving}>
						{saving && <Loader2 className="size-4 animate-spin" />}
						{isEditing ? "Salvar alteracoes" : "Criar artigo"}
					</Button>
				</div>
			</form>

			<Dialog open={mediaOpen} onOpenChange={setMediaOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Inserir midia</DialogTitle>
					</DialogHeader>
					<MediaUpload onUpload={handleMediaInsert} />
				</DialogContent>
			</Dialog>
		</>
	);
}
