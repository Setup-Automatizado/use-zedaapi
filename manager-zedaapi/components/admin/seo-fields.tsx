"use client";

import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type SeoFieldKey = "seoTitle" | "seoDescription" | "seoKeywords";

interface SeoFieldsProps {
	seoTitle: string;
	seoDescription: string;
	seoKeywords: string;
	onChange: (field: SeoFieldKey, value: string) => void;
	previewUrl?: string;
	className?: string;
}

function CharBadge({
	count,
	warn,
	max,
}: {
	count: number;
	warn: number;
	max: number;
}) {
	const variant =
		count > max ? "destructive" : count > warn ? "outline" : "secondary";
	return (
		<Badge variant={variant} className="tabular-nums">
			{count}/{max}
		</Badge>
	);
}

export function SeoFields({
	seoTitle,
	seoDescription,
	seoKeywords,
	onChange,
	previewUrl = "https://zedaapi.com/blog/...",
	className,
}: SeoFieldsProps) {
	return (
		<div className={cn("space-y-4", className)}>
			<div className="space-y-1.5">
				<div className="flex items-center justify-between">
					<Label>Titulo SEO</Label>
					<CharBadge count={seoTitle.length} warn={50} max={60} />
				</div>
				<Input
					value={seoTitle}
					onChange={(e) => onChange("seoTitle", e.target.value)}
					placeholder="Titulo otimizado para motores de busca"
				/>
			</div>

			<div className="space-y-1.5">
				<div className="flex items-center justify-between">
					<Label>Descricao SEO</Label>
					<CharBadge
						count={seoDescription.length}
						warn={140}
						max={160}
					/>
				</div>
				<Textarea
					value={seoDescription}
					onChange={(e) => onChange("seoDescription", e.target.value)}
					placeholder="Descricao que aparece nos resultados de busca"
					rows={2}
					className="min-h-0"
				/>
			</div>

			<div className="space-y-1.5">
				<Label>Palavras-chave</Label>
				<Input
					value={seoKeywords}
					onChange={(e) => onChange("seoKeywords", e.target.value)}
					placeholder="whatsapp, api, automacao (separadas por virgula)"
				/>
			</div>

			<div className="space-y-2">
				<Label className="text-muted-foreground">
					Preview do Google
				</Label>
				<Card size="sm" className="border-dashed">
					<CardContent className="space-y-1">
						<p className="line-clamp-1 text-base font-medium text-blue-600 dark:text-blue-400">
							{seoTitle || "Titulo da pagina"}
						</p>
						<p className="line-clamp-1 text-xs text-green-700 dark:text-green-500">
							{previewUrl}
						</p>
						<p className="line-clamp-2 text-sm text-muted-foreground">
							{seoDescription ||
								"A descricao do conteudo aparecera aqui..."}
						</p>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
