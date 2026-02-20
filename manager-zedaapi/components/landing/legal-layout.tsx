import Link from "next/link";
import { ArrowLeftIcon, PrinterIcon, CalendarIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface LegalLayoutProps {
	title: string;
	lastUpdated: string;
	children: React.ReactNode;
}

const legalPages = [
	{ label: "Termos de Uso", href: "/termos-de-uso" },
	{ label: "Privacidade", href: "/politica-de-privacidade" },
	{ label: "Cookies", href: "/politica-de-cookies" },
	{ label: "LGPD", href: "/lgpd" },
	{ label: "Exclusão de Dados", href: "/exclusao-de-dados" },
] as const;

export function LegalLayout({
	title,
	lastUpdated,
	children,
}: LegalLayoutProps) {
	return (
		<div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8 lg:py-20">
			{/* Back link */}
			<Link
				href="/"
				className="mb-8 inline-flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors duration-200 hover:text-foreground"
			>
				<ArrowLeftIcon className="size-4" />
				Voltar ao início
			</Link>

			<div className="flex gap-12 lg:gap-16">
				{/* Sidebar navigation - desktop only */}
				<aside className="hidden w-56 shrink-0 lg:block">
					<div className="sticky top-24">
						<p className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
							Legal
						</p>
						<nav className="flex flex-col gap-1">
							{legalPages.map((page) => (
								<Link
									key={page.href}
									href={page.href}
									className={cn(
										"rounded-lg px-3 py-2 text-sm transition-colors duration-200 hover:bg-muted hover:text-foreground",
										"text-muted-foreground",
									)}
								>
									{page.label}
								</Link>
							))}
						</nav>
					</div>
				</aside>

				{/* Content */}
				<article className="min-w-0 flex-1 max-w-4xl">
					{/* Header */}
					<header className="mb-10">
						<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
							{title}
						</h1>
						<div className="mt-4 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
							<span className="inline-flex items-center gap-1.5">
								<CalendarIcon className="size-3.5" />
								Última atualização: {lastUpdated}
							</span>
							<button
								type="button"
								className="inline-flex items-center gap-1.5 transition-colors duration-200 hover:text-foreground print:hidden"
								onClick={undefined}
							>
								<PrinterIcon className="size-3.5" />
								Imprimir
							</button>
						</div>
					</header>

					{/* Legal content with prose styling */}
					<div
						className={cn(
							"legal-content",
							"text-sm leading-7 text-muted-foreground sm:text-base sm:leading-8",
							"[&_h2]:mt-12 [&_h2]:mb-4 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:text-foreground [&_h2]:tracking-tight sm:[&_h2]:text-2xl",
							"[&_h3]:mt-8 [&_h3]:mb-3 [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:text-foreground",
							"[&_h4]:mt-6 [&_h4]:mb-2 [&_h4]:text-base [&_h4]:font-medium [&_h4]:text-foreground",
							"[&_p]:mb-4",
							"[&_ul]:mb-4 [&_ul]:ml-6 [&_ul]:list-disc [&_ul]:space-y-2",
							"[&_ol]:mb-4 [&_ol]:ml-6 [&_ol]:list-decimal [&_ol]:space-y-2",
							"[&_li]:pl-1",
							"[&_a]:text-primary [&_a]:underline [&_a]:underline-offset-4 [&_a]:transition-colors [&_a]:duration-200 hover:[&_a]:text-primary/80",
							"[&_strong]:font-semibold [&_strong]:text-foreground",
							"[&_table]:mb-6 [&_table]:w-full [&_table]:border-collapse [&_table]:text-sm",
							"[&_th]:border [&_th]:border-border [&_th]:bg-muted/50 [&_th]:px-4 [&_th]:py-2 [&_th]:text-left [&_th]:font-medium [&_th]:text-foreground",
							"[&_td]:border [&_td]:border-border [&_td]:px-4 [&_td]:py-2",
							"[&_blockquote]:my-6 [&_blockquote]:border-l-2 [&_blockquote]:border-primary/30 [&_blockquote]:pl-4 [&_blockquote]:italic",
							"[&_hr]:my-8 [&_hr]:border-border",
						)}
					>
						{children}
					</div>

					{/* Mobile legal nav */}
					<div className="mt-16 border-t border-border pt-8 lg:hidden">
						<p className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
							Outros documentos legais
						</p>
						<nav className="flex flex-wrap gap-2">
							{legalPages.map((page) => (
								<Link
									key={page.href}
									href={page.href}
									className="rounded-lg border border-border px-3 py-1.5 text-sm text-muted-foreground transition-colors duration-200 hover:bg-muted hover:text-foreground"
								>
									{page.label}
								</Link>
							))}
						</nav>
					</div>
				</article>
			</div>
		</div>
	);
}
