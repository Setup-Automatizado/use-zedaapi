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
import { Loader2 } from "lucide-react";
import {
	createGlossaryTerm,
	updateGlossaryTerm,
} from "@/server/actions/content";

interface GlossarioFormClientProps {
	initialData?: {
		id: string;
		term: string;
		slug: string;
		definition: string;
		content: string | null;
		seoTitle: string | null;
		seoDescription: string | null;
		status: string;
		relatedSlugs: string[] | null;
	};
}

export function GlossarioFormClient({ initialData }: GlossarioFormClientProps) {
	const router = useRouter();
	const [loading, setLoading] = useState(false);

	const [term, setTerm] = useState(initialData?.term ?? "");
	const [slug, setSlug] = useState(initialData?.slug ?? "");
	const [definition, setDefinition] = useState(initialData?.definition ?? "");
	const [content, setContent] = useState(initialData?.content ?? "");
	const [status, setStatus] = useState(initialData?.status ?? "draft");
	const [relatedSlugsText, setRelatedSlugsText] = useState(
		initialData?.relatedSlugs?.join(", ") ?? "",
	);
	const [seoTitle, setSeoTitle] = useState(initialData?.seoTitle ?? "");
	const [seoDescription, setSeoDescription] = useState(
		initialData?.seoDescription ?? "",
	);
	const [seoKeywords, setSeoKeywords] = useState("");

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();

		if (!term.trim()) {
			toast.error("O termo e obrigatorio");
			return;
		}
		if (!definition.trim()) {
			toast.error("A definicao e obrigatoria");
			return;
		}

		setLoading(true);

		const relatedSlugs = relatedSlugsText
			.split(",")
			.map((s) => s.trim())
			.filter(Boolean);

		const formData = {
			term: term.trim(),
			definition: definition.trim(),
			content: content.trim() || undefined,
			seoTitle: seoTitle.trim() || undefined,
			seoDescription: seoDescription.trim() || undefined,
			status,
			relatedSlugs: relatedSlugs.length > 0 ? relatedSlugs : undefined,
		};

		const res = initialData
			? await updateGlossaryTerm(initialData.id, formData)
			: await createGlossaryTerm(formData);

		setLoading(false);

		if (res.success) {
			toast.success(
				initialData
					? "Termo atualizado com sucesso"
					: "Termo criado com sucesso",
			);
			router.push("/admin/glossario");
		} else {
			toast.error(res.error ?? "Erro ao salvar termo");
		}
	}

	function handleSeoChange(
		field: "seoTitle" | "seoDescription" | "seoKeywords",
		value: string,
	) {
		if (field === "seoTitle") setSeoTitle(value);
		else if (field === "seoDescription") setSeoDescription(value);
		else setSeoKeywords(value);
	}

	return (
		<form onSubmit={handleSubmit} className="space-y-6">
			<Card>
				<CardHeader>
					<CardTitle>Informacoes basicas</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="space-y-1.5">
						<Label htmlFor="term">Termo</Label>
						<Input
							id="term"
							value={term}
							onChange={(e) => setTerm(e.target.value)}
							placeholder="Ex: API REST"
						/>
					</div>

					<SlugInput
						value={slug}
						onChange={setSlug}
						sourceValue={term}
						prefix="/glossario"
					/>

					<div className="space-y-1.5">
						<Label htmlFor="status">Status</Label>
						<Select value={status} onValueChange={setStatus}>
							<SelectTrigger id="status">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="draft">Rascunho</SelectItem>
								<SelectItem value="published">
									Publicado
								</SelectItem>
							</SelectContent>
						</Select>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>Definicao</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="space-y-1.5">
						<Label htmlFor="definition">Definicao curta</Label>
						<Textarea
							id="definition"
							value={definition}
							onChange={(e) => setDefinition(e.target.value)}
							placeholder="Uma definicao breve que aparece no indice do glossario"
							rows={3}
							className="min-h-0"
						/>
					</div>

					<div className="space-y-1.5">
						<Label>Conteudo completo</Label>
						<ContentEditor value={content} onChange={setContent} />
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>Termos relacionados</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="space-y-1.5">
						<Label htmlFor="relatedSlugs">Slugs relacionados</Label>
						<Input
							id="relatedSlugs"
							value={relatedSlugsText}
							onChange={(e) =>
								setRelatedSlugsText(e.target.value)
							}
							placeholder="api-rest, webhook, websocket"
						/>
						<p className="text-xs text-muted-foreground">
							Slugs de termos relacionados, separados por virgula
						</p>
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
						previewUrl={`https://zedaapi.com/glossario/${slug || "..."}`}
					/>
				</CardContent>
			</Card>

			<div className="flex justify-end gap-3">
				<Button
					type="button"
					variant="outline"
					onClick={() => router.push("/admin/glossario")}
				>
					Cancelar
				</Button>
				<Button type="submit" disabled={loading}>
					{loading && <Loader2 className="size-4 animate-spin" />}
					{initialData ? "Salvar alteracoes" : "Criar termo"}
				</Button>
			</div>
		</form>
	);
}
