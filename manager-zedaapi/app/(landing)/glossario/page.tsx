import { Search } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { db } from "@/lib/db";

export const metadata: Metadata = {
	title: "Glossario - Zé da API",
	description:
		"Glossario de termos sobre WhatsApp API, automacao de mensagens, webhooks e integracoes. Entenda os conceitos usados na plataforma.",
	openGraph: {
		title: "Glossario - Zé da API",
		description:
			"Glossario de termos sobre WhatsApp API e automacao de mensagens.",
		url: "https://zedaapi.com/glossario",
	},
	alternates: {
		canonical: "https://zedaapi.com/glossario",
	},
};

export const dynamic = "force-dynamic";

const ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ".split("");

export default async function GlossarioPage() {
	const terms = await db.glossaryTerm.findMany({
		where: { status: "published" },
		orderBy: { term: "asc" },
		select: { term: true, slug: true, definition: true },
	});

	// Group by first letter
	const grouped: Record<string, typeof terms> = {};
	for (const term of terms) {
		const letter = term.term.charAt(0).toUpperCase();
		if (!grouped[letter]) grouped[letter] = [];
		grouped[letter].push(term);
	}

	const activeLetters = new Set(Object.keys(grouped));

	const jsonLd = {
		"@context": "https://schema.org",
		"@type": "DefinedTermSet",
		name: "Glossario Zé da API",
		description:
			"Glossario de termos sobre WhatsApp API e automacao de mensagens.",
		url: "https://zedaapi.com/glossario",
		hasDefinedTerm: terms.map((t) => ({
			"@type": "DefinedTerm",
			name: t.term,
			description: t.definition,
			url: `https://zedaapi.com/glossario/${t.slug}`,
		})),
	};

	return (
		<div>
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify(jsonLd),
				}}
			/>

			{/* Hero */}
			<section className="border-b border-border bg-muted/30 py-16 sm:py-20">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<div className="mx-auto max-w-2xl text-center">
						<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
							Glossario
						</h1>
						<p className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
							Termos e conceitos sobre WhatsApp API, automacao e
							integracoes.
						</p>
					</div>
				</div>
			</section>

			{/* Content */}
			<section className="py-12 sm:py-16">
				<div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
					{/* Alphabet nav */}
					<nav className="mb-8 flex flex-wrap gap-1">
						{ALPHABET.map((letter) => (
							<a
								key={letter}
								href={
									activeLetters.has(letter)
										? `#${letter}`
										: undefined
								}
								className={`flex size-9 items-center justify-center rounded-lg text-sm font-medium transition-colors ${
									activeLetters.has(letter)
										? "bg-primary/10 text-primary hover:bg-primary hover:text-primary-foreground"
										: "text-muted-foreground/40 cursor-default"
								}`}
							>
								{letter}
							</a>
						))}
					</nav>

					{/* Terms grouped by letter */}
					{terms.length === 0 ? (
						<div className="flex min-h-[300px] items-center justify-center rounded-2xl border border-dashed border-border">
							<div className="text-center">
								<Search className="mx-auto size-8 text-muted-foreground/50" />
								<p className="mt-3 text-sm font-medium">
									Glossario em construcao
								</p>
								<p className="mt-1 text-xs text-muted-foreground">
									Termos serao adicionados em breve.
								</p>
							</div>
						</div>
					) : (
						<div className="space-y-10">
							{ALPHABET.filter((l) => activeLetters.has(l)).map(
								(letter) => (
									<div key={letter} id={letter}>
										<h2 className="mb-4 border-b border-border pb-2 text-2xl font-bold text-foreground">
											{letter}
										</h2>
										<div className="space-y-3">
											{grouped[letter]?.map((term) => (
												<Link
													key={term.slug}
													href={`/glossario/${term.slug}`}
													className="group block rounded-xl border border-border p-4 transition-all duration-150 hover:-translate-y-0.5 hover:shadow-md"
												>
													<h3 className="font-semibold text-foreground group-hover:text-primary transition-colors">
														{term.term}
													</h3>
													<p className="mt-1 text-sm leading-relaxed text-muted-foreground line-clamp-2">
														{term.definition}
													</p>
												</Link>
											))}
										</div>
									</div>
								),
							)}
						</div>
					)}
				</div>
			</section>
		</div>
	);
}
