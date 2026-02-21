"use client";

import { useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
} from "@/components/ui/dialog";
import { ContentEditor } from "@/components/admin/content-editor";
import { SlugInput } from "@/components/admin/slug-input";
import { SeoFields } from "@/components/admin/seo-fields";
import { MediaUpload } from "@/components/admin/media-upload";
import { Loader2, X, ImagePlus, Plus } from "lucide-react";
import {
	createBlogPost,
	updateBlogPost,
	createBlogTag,
} from "@/server/actions/blog";

interface BlogFormClientProps {
	categories: Array<{ id: string; name: string }>;
	tags: Array<{ id: string; name: string; slug: string }>;
	initialData?: {
		id: string;
		title: string;
		slug: string;
		content: string;
		excerpt: string | null;
		coverImageUrl: string | null;
		categoryId: string | null;
		seoTitle: string | null;
		seoDescription: string | null;
		seoKeywords: string | null;
		status: string;
		tags?: Array<{ tag: { id: string; name: string } }>;
		media?: Array<{
			id: string;
			url: string;
			type: string;
			filename: string;
			alt: string | null;
			caption: string | null;
		}>;
	};
}

const NO_CATEGORY = "__none__";

export function BlogFormClient({
	categories,
	tags: initialTags,
	initialData,
}: BlogFormClientProps) {
	const router = useRouter();
	const isEditing = !!initialData;

	const [title, setTitle] = useState(initialData?.title ?? "");
	const [slug, setSlug] = useState(initialData?.slug ?? "");
	const [content, setContent] = useState(initialData?.content ?? "");
	const [excerpt, setExcerpt] = useState(initialData?.excerpt ?? "");
	const [coverImageUrl, setCoverImageUrl] = useState(
		initialData?.coverImageUrl ?? "",
	);
	const [categoryId, setCategoryId] = useState(
		initialData?.categoryId ?? NO_CATEGORY,
	);
	const [status, setStatus] = useState(initialData?.status ?? "draft");
	const [seoTitle, setSeoTitle] = useState(initialData?.seoTitle ?? "");
	const [seoDescription, setSeoDescription] = useState(
		initialData?.seoDescription ?? "",
	);
	const [seoKeywords, setSeoKeywords] = useState(
		initialData?.seoKeywords ?? "",
	);

	const [selectedTagIds, setSelectedTagIds] = useState<Set<string>>(
		new Set(initialData?.tags?.map((t) => t.tag.id) ?? []),
	);
	const [availableTags, setAvailableTags] = useState(initialTags);
	const [newTagName, setNewTagName] = useState("");
	const [creatingTag, setCreatingTag] = useState(false);

	const [mediaDialogOpen, setMediaDialogOpen] = useState(false);
	const [submitting, setSubmitting] = useState(false);

	const toggleTag = useCallback((tagId: string) => {
		setSelectedTagIds((prev) => {
			const next = new Set(prev);
			if (next.has(tagId)) {
				next.delete(tagId);
			} else {
				next.add(tagId);
			}
			return next;
		});
	}, []);

	const handleCreateTag = useCallback(async () => {
		const name = newTagName.trim();
		if (!name) return;

		setCreatingTag(true);
		const res = await createBlogTag({ name });
		if (res.success) {
			toast.success("Tag criada com sucesso");
			setNewTagName("");
			// Refetch tags by re-importing action
			const { getAdminBlogTags } = await import("@/server/actions/blog");
			const tagRes = await getAdminBlogTags();
			if (tagRes.success && tagRes.data) {
				const newTags = tagRes.data as Array<{
					id: string;
					name: string;
					slug: string;
				}>;
				setAvailableTags(newTags);
				// Auto-select the newly created tag
				const created = newTags.find(
					(t) => t.name.toLowerCase() === name.toLowerCase(),
				);
				if (created) {
					setSelectedTagIds((prev) => new Set([...prev, created.id]));
				}
			}
		} else {
			toast.error(res.error ?? "Erro ao criar tag");
		}
		setCreatingTag(false);
	}, [newTagName]);

	const handleInsertMedia = useCallback(() => {
		setMediaDialogOpen(true);
	}, []);

	const handleMediaInsert = useCallback(
		(media: { url: string; filename: string; type: string }) => {
			let markdown: string;
			if (media.type.startsWith("image/")) {
				markdown = `\n![${media.filename}](${media.url})\n`;
			} else if (
				media.type.startsWith("video/") ||
				media.type === "video/youtube"
			) {
				markdown = `\n[${media.filename}](${media.url})\n`;
			} else {
				markdown = `\n[${media.filename}](${media.url})\n`;
			}
			setContent((prev) => prev + markdown);
			setMediaDialogOpen(false);
		},
		[],
	);

	const handleSeoChange = useCallback(
		(
			field: "seoTitle" | "seoDescription" | "seoKeywords",
			value: string,
		) => {
			switch (field) {
				case "seoTitle":
					setSeoTitle(value);
					break;
				case "seoDescription":
					setSeoDescription(value);
					break;
				case "seoKeywords":
					setSeoKeywords(value);
					break;
			}
		},
		[],
	);

	const handleSubmit = useCallback(
		async (e: React.FormEvent) => {
			e.preventDefault();

			if (!title.trim()) {
				toast.error("O titulo e obrigatorio");
				return;
			}
			if (!content.trim()) {
				toast.error("O conteudo e obrigatorio");
				return;
			}

			setSubmitting(true);

			const formData = {
				title: title.trim(),
				content,
				excerpt: excerpt.trim() || undefined,
				coverImageUrl: coverImageUrl.trim() || undefined,
				categoryId: categoryId === NO_CATEGORY ? undefined : categoryId,
				seoTitle: seoTitle.trim() || undefined,
				seoDescription: seoDescription.trim() || undefined,
				seoKeywords: seoKeywords.trim() || undefined,
				status,
				tagIds: Array.from(selectedTagIds),
			};

			if (isEditing && initialData) {
				const res = await updateBlogPost(initialData.id, formData);
				if (res.success) {
					toast.success("Post atualizado com sucesso");
					router.push("/admin/blog");
				} else {
					toast.error(res.error ?? "Erro ao atualizar post");
				}
			} else {
				const res = await createBlogPost(formData);
				if (res.success) {
					toast.success("Post criado com sucesso");
					router.push("/admin/blog");
				} else {
					toast.error(res.error ?? "Erro ao criar post");
				}
			}

			setSubmitting(false);
		},
		[
			title,
			content,
			excerpt,
			coverImageUrl,
			categoryId,
			seoTitle,
			seoDescription,
			seoKeywords,
			status,
			selectedTagIds,
			isEditing,
			initialData,
			router,
		],
	);

	return (
		<form onSubmit={handleSubmit} className="space-y-6">
			<div className="grid gap-6 lg:grid-cols-3">
				<div className="space-y-6 lg:col-span-2">
					{/* Informacoes basicas */}
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
									placeholder="Titulo do post"
								/>
							</div>

							<SlugInput
								value={slug}
								onChange={setSlug}
								sourceValue={title}
								prefix="/blog"
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
											<SelectItem value={NO_CATEGORY}>
												Sem categoria
											</SelectItem>
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
						</CardContent>
					</Card>

					{/* Conteudo */}
					<Card>
						<CardHeader>
							<CardTitle>Conteudo</CardTitle>
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="space-y-1.5">
								<Label htmlFor="excerpt">Resumo do post</Label>
								<Textarea
									id="excerpt"
									value={excerpt}
									onChange={(e) => setExcerpt(e.target.value)}
									placeholder="Um breve resumo que aparece nas listagens..."
									rows={3}
									className="min-h-0"
								/>
							</div>

							<div className="space-y-1.5">
								<Label>Conteudo (Markdown)</Label>
								<ContentEditor
									value={content}
									onChange={setContent}
									onInsertMedia={handleInsertMedia}
								/>
							</div>
						</CardContent>
					</Card>

					{/* SEO */}
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
								previewUrl={
									slug
										? `https://zedaapi.com/blog/${slug}`
										: "https://zedaapi.com/blog/..."
								}
							/>
						</CardContent>
					</Card>
				</div>

				<div className="space-y-6">
					{/* Tags */}
					<Card>
						<CardHeader>
							<CardTitle>Tags</CardTitle>
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="flex flex-wrap gap-2">
								{availableTags.map((tag) => (
									<Badge
										key={tag.id}
										variant={
											selectedTagIds.has(tag.id)
												? "default"
												: "outline"
										}
										className="cursor-pointer transition-colors"
										onClick={() => toggleTag(tag.id)}
									>
										{tag.name}
										{selectedTagIds.has(tag.id) && (
											<X className="ml-1 size-3" />
										)}
									</Badge>
								))}
								{availableTags.length === 0 && (
									<p className="text-sm text-muted-foreground">
										Nenhuma tag disponivel
									</p>
								)}
							</div>

							<div className="flex gap-2">
								<Input
									value={newTagName}
									onChange={(e) =>
										setNewTagName(e.target.value)
									}
									placeholder="Nova tag..."
									className="flex-1"
									onKeyDown={(e) => {
										if (e.key === "Enter") {
											e.preventDefault();
											void handleCreateTag();
										}
									}}
								/>
								<Button
									type="button"
									variant="outline"
									size="icon"
									onClick={() => void handleCreateTag()}
									disabled={creatingTag || !newTagName.trim()}
								>
									{creatingTag ? (
										<Loader2 className="size-4 animate-spin" />
									) : (
										<Plus className="size-4" />
									)}
								</Button>
							</div>
						</CardContent>
					</Card>

					{/* Imagem de capa */}
					<Card>
						<CardHeader>
							<CardTitle>Imagem de capa</CardTitle>
						</CardHeader>
						<CardContent className="space-y-4">
							{coverImageUrl ? (
								<div className="relative">
									<img
										src={coverImageUrl}
										alt="Capa do post"
										className="aspect-video w-full rounded-lg border object-cover"
									/>
									<Button
										type="button"
										variant="destructive"
										size="icon-xs"
										className="absolute right-2 top-2"
										onClick={() => setCoverImageUrl("")}
										aria-label="Remover imagem de capa"
									>
										<X className="size-3.5" />
									</Button>
								</div>
							) : (
								<MediaUpload
									onUpload={(media) =>
										setCoverImageUrl(media.url)
									}
									accept="image/*"
								/>
							)}
						</CardContent>
					</Card>

					{/* Acoes */}
					<Card>
						<CardContent className="pt-6">
							<Button
								type="submit"
								className="w-full"
								disabled={submitting}
							>
								{submitting && (
									<Loader2 className="size-4 animate-spin" />
								)}
								{isEditing ? "Salvar alteracoes" : "Criar post"}
							</Button>
						</CardContent>
					</Card>
				</div>
			</div>

			{/* Dialog para inserir midia no conteudo */}
			<Dialog open={mediaDialogOpen} onOpenChange={setMediaDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							<div className="flex items-center gap-2">
								<ImagePlus className="size-5" />
								Inserir midia
							</div>
						</DialogTitle>
					</DialogHeader>
					<MediaUpload onUpload={handleMediaInsert} />
				</DialogContent>
			</Dialog>
		</form>
	);
}
