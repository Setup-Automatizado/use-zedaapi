import type { Metadata } from "next";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { PostContent } from "@/components/blog/post-content";
import { db } from "@/lib/db";

interface GlossaryTermPageProps {
	params: Promise<{ slug: string }>;
}

export async function generateMetadata({
	params,
}: GlossaryTermPageProps): Promise<Metadata> {
	const { slug } = await params;

	const term = await db.glossaryTerm.findFirst({
		where: { slug, status: "published" },
		select: {
			term: true,
			definition: true,
			seoTitle: true,
			seoDescription: true,
			slug: true,
		},
	});

	if (!term) return { title: "Termo nao encontrado" };

	const title = term.seoTitle || `${term.term} - Glossario`;
	const description = term.seoDescription || term.definition;

	return {
		title: `${title} - Zé da API`,
		description,
		alternates: {
			canonical: `https://zedaapi.com/glossario/${term.slug}`,
		},
	};
}

export default async function GlossaryTermPage({
	params,
}: GlossaryTermPageProps) {
	const { slug } = await params;

	const term = await db.glossaryTerm.findFirst({
		where: { slug, status: "published" },
	});

	if (!term) notFound();

	// Get related terms
	const relatedSlugs = (term.relatedSlugs as string[] | null) ?? [];
	const relatedTerms =
		relatedSlugs.length > 0
			? await db.glossaryTerm.findMany({
					where: {
						slug: { in: relatedSlugs },
						status: "published",
					},
					select: { term: true, slug: true, definition: true },
				})
			: [];

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "DefinedTerm",
		name: term.term,
		description: term.definition,
		url: `https://zedaapi.com/glossario/${term.slug}`,
		inDefinedTermSet: {
			"@type": "DefinedTermSet",
			name: "Glossario Zé da API",
			url: "https://zedaapi.com/glossario",
		},
	};

	return (
		<div>
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify(jsonLd),
				}}
			/>

			<section className="py-10 sm:py-14">
				<div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
					{/* Back */}
					<Link
						href="/glossario"
						className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
					>
						<ArrowLeft className="size-3.5" />
						Voltar ao glossario
					</Link>

					<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
						{term.term}
					</h1>

					<p className="mt-4 text-lg leading-relaxed text-muted-foreground">
						{term.definition}
					</p>

					{term.content && (
						<div className="mt-8 border-t border-border pt-8">
							<PostContent content={term.content} />
						</div>
					)}

					{/* Related terms */}
					{relatedTerms.length > 0 && (
						<div className="mt-10 rounded-2xl border border-border bg-muted/30 p-6">
							<h3 className="text-sm font-semibold text-foreground">
								Termos relacionados
							</h3>
							<div className="mt-3 flex flex-wrap gap-2">
								{relatedTerms.map((rt) => (
									<Link
										key={rt.slug}
										href={`/glossario/${rt.slug}`}
										className="rounded-lg border border-border bg-card px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-accent hover:text-primary"
									>
										{rt.term}
									</Link>
								))}
							</div>
						</div>
					)}
				</div>
			</section>
		</div>
	);
}
