import type { Metadata } from "next";
import Link from "next/link";
import {
	Search,
	HelpCircle,
	BookOpen,
	Settings,
	Zap,
	CreditCard,
	Shield,
	MessageSquare,
} from "lucide-react";
import { db } from "@/lib/db";

export const metadata: Metadata = {
	title: "Central de Suporte - Zé da API",
	description:
		"Encontre respostas para suas duvidas sobre a Zé da API. Artigos de suporte, tutoriais e guias para usar a plataforma.",
	openGraph: {
		title: "Central de Suporte - Zé da API",
		description:
			"Central de suporte com artigos e tutoriais para a Zé da API.",
		url: "https://zedaapi.com/suporte",
	},
	alternates: {
		canonical: "https://zedaapi.com/suporte",
	},
};

export const dynamic = "force-dynamic";

const iconMap: Record<string, typeof HelpCircle> = {
	"help-circle": HelpCircle,
	"book-open": BookOpen,
	settings: Settings,
	zap: Zap,
	"credit-card": CreditCard,
	shield: Shield,
	"message-square": MessageSquare,
};

export default async function SuportePage() {
	const categories = await db.supportCategory.findMany({
		where: {
			articles: { some: { status: "published" } },
		},
		include: {
			_count: {
				select: {
					articles: { where: { status: "published" } },
				},
			},
		},
		orderBy: { sortOrder: "asc" },
	});

	const popularArticles = await db.supportArticle.findMany({
		where: { status: "published" },
		include: { category: { select: { name: true } } },
		orderBy: { viewCount: "desc" },
		take: 6,
	});

	return (
		<div>
			{/* Hero */}
			<section className="border-b border-border bg-muted/30 py-16 sm:py-20">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<div className="mx-auto max-w-2xl text-center">
						<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
							Como podemos ajudar?
						</h1>
						<p className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
							Encontre respostas, tutoriais e guias para usar a Ze
							da API.
						</p>
					</div>
				</div>
			</section>

			{/* Categories */}
			<section className="py-12 sm:py-16">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					{categories.length > 0 && (
						<>
							<h2 className="mb-6 text-xl font-semibold text-foreground">
								Categorias
							</h2>
							<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
								{categories.map((cat) => {
									const Icon =
										iconMap[cat.icon ?? "help-circle"] ??
										HelpCircle;
									return (
										<Link
											key={cat.id}
											href={`/suporte?categoria=${cat.slug}`}
											className="group rounded-2xl border border-border bg-card p-6 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-lg"
										>
											<div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
												<Icon className="size-5" />
											</div>
											<h3 className="mt-4 font-semibold text-foreground">
												{cat.name}
											</h3>
											{cat.description && (
												<p className="mt-1 text-sm text-muted-foreground line-clamp-2">
													{cat.description}
												</p>
											)}
											<p className="mt-2 text-xs text-muted-foreground">
												{cat._count.articles}{" "}
												{cat._count.articles === 1
													? "artigo"
													: "artigos"}
											</p>
										</Link>
									);
								})}
							</div>
						</>
					)}

					{/* Popular articles */}
					{popularArticles.length > 0 && (
						<div className="mt-12">
							<h2 className="mb-6 text-xl font-semibold text-foreground">
								Artigos populares
							</h2>
							<div className="grid gap-3 sm:grid-cols-2">
								{popularArticles.map((article) => (
									<Link
										key={article.id}
										href={`/suporte/${article.slug}`}
										className="group flex items-start gap-3 rounded-xl border border-border p-4 transition-colors duration-150 hover:bg-accent/50"
									>
										<HelpCircle className="mt-0.5 size-4 shrink-0 text-muted-foreground group-hover:text-primary" />
										<div>
											<p className="text-sm font-medium text-foreground group-hover:text-primary">
												{article.title}
											</p>
											{article.excerpt && (
												<p className="mt-0.5 text-xs text-muted-foreground line-clamp-1">
													{article.excerpt}
												</p>
											)}
											<p className="mt-1 text-xs text-muted-foreground">
												{article.category.name}
											</p>
										</div>
									</Link>
								))}
							</div>
						</div>
					)}

					{categories.length === 0 &&
						popularArticles.length === 0 && (
							<div className="flex min-h-[300px] items-center justify-center rounded-2xl border border-dashed border-border">
								<div className="text-center">
									<Search className="mx-auto size-8 text-muted-foreground/50" />
									<p className="mt-3 text-sm font-medium">
										Central de suporte em construcao
									</p>
									<p className="mt-1 text-xs text-muted-foreground">
										Artigos serao adicionados em breve.
									</p>
								</div>
							</div>
						)}

					{/* Contact CTA */}
					<div className="mt-12 rounded-2xl border border-border bg-muted/30 p-8 text-center">
						<MessageSquare className="mx-auto size-8 text-primary" />
						<h3 className="mt-3 text-lg font-semibold text-foreground">
							Nao encontrou o que procura?
						</h3>
						<p className="mt-2 text-sm text-muted-foreground">
							Nossa equipe esta pronta para ajudar.
						</p>
						<Link
							href="/contato"
							className="mt-4 inline-flex items-center rounded-lg bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
						>
							Fale conosco
						</Link>
					</div>
				</div>
			</section>
		</div>
	);
}
