"use client";

import Link from "next/link";
import Image from "next/image";
import {
	MailIcon,
	PhoneIcon,
	HeadphonesIcon,
	GithubIcon,
	LinkedinIcon,
	TwitterIcon,
	ArrowUpRightIcon,
	MapPinIcon,
	ShieldCheckIcon,
	ArrowRightIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";

// ── Link data ──────────────────────────

type FooterLink = {
	label: string;
	href: string;
	external?: boolean;
	badge?: string;
};

const productLinks: FooterLink[] = [
	{ label: "Recursos", href: "#recursos" },
	{ label: "Preços", href: "#precos" },
	{ label: "Integrações", href: "#integracoes" },
	{
		label: "Documentação",
		href: "https://api.zedaapi.com/docs",
		external: true,
	},
	{
		label: "API Reference",
		href: "https://api.zedaapi.com/docs",
		external: true,
	},
	{ label: "Status", href: "/status" },
];

const resourceLinks: FooterLink[] = [
	{ label: "Blog", href: "/blog" },
	{
		label: "Changelog",
		href: "https://github.com/Setup-Automatizado/use-zedaapi/releases",
		external: true,
	},
	{
		label: "Guias de Integração",
		href: "https://api.zedaapi.com/docs",
		external: true,
	},
	{
		label: "Postman Collection",
		href: "https://api.zedaapi.com/docs",
		external: true,
	},
	{
		label: "Node n8n",
		href: "https://www.npmjs.com/package/@setup-automatizado/n8n-nodes-zedaapi",
		external: true,
		badge: "Novo",
	},
	{ label: "Suporte", href: "/suporte" },
];

const legalLinks: FooterLink[] = [
	{ label: "Termos de Uso", href: "/termos-de-uso" },
	{ label: "Política de Privacidade", href: "/politica-de-privacidade" },
	{ label: "Política de Cookies", href: "/politica-de-cookies" },
	{ label: "LGPD", href: "/lgpd" },
	{ label: "Exclusão de Dados", href: "/exclusao-de-dados" },
];

const contactInfo = [
	{
		icon: MailIcon,
		label: "contato@zedaapi.com",
		href: "mailto:contato@zedaapi.com",
	},
	{
		icon: PhoneIcon,
		label: "+55 21 97153-2700",
		href: "https://wa.me/5521971532700",
	},
	{
		icon: HeadphonesIcon,
		label: "suporte@zedaapi.com",
		href: "mailto:suporte@zedaapi.com",
	},
] as const;

const socialLinks = [
	{
		icon: GithubIcon,
		label: "GitHub",
		href: "https://github.com/Setup-Automatizado/use-zedaapi",
	},
	{
		icon: LinkedinIcon,
		label: "LinkedIn",
		href: "https://linkedin.com/company/zedaapi",
	},
	{
		icon: TwitterIcon,
		label: "Twitter",
		href: "https://twitter.com/zedaapi",
	},
] as const;

// ── Helper: render a link ──────────────────────────

function FooterLinkItem({ link }: { link: FooterLink }) {
	const isExternal = link.external || link.href.startsWith("http");
	const isAnchor = link.href.startsWith("#");

	const className =
		"group/link inline-flex items-center gap-1 rounded-sm text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background";

	if (isAnchor) {
		return (
			<a href={link.href} className={className}>
				{link.label}
			</a>
		);
	}

	if (isExternal) {
		return (
			<a
				href={link.href}
				target="_blank"
				rel="noopener noreferrer"
				className={className}
			>
				{link.label}
				<ArrowUpRightIcon className="size-3 opacity-0 -translate-y-0.5 translate-x-[-2px] transition-all duration-200 group-hover/link:opacity-60 group-hover/link:translate-y-0 group-hover/link:translate-x-0" />
				{link.badge && (
					<span className="ml-1 rounded-full bg-primary/10 px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary">
						{link.badge}
					</span>
				)}
			</a>
		);
	}

	return (
		<Link href={link.href} className={className}>
			{link.label}
			{link.badge && (
				<span className="ml-1 rounded-full bg-primary/10 px-1.5 py-0.5 text-[10px] font-semibold leading-none text-primary">
					{link.badge}
				</span>
			)}
		</Link>
	);
}

// ── Component ──────────────────────────

export function Footer() {
	return (
		<footer className="relative overflow-hidden border-t border-border/40">
			{/* Background layers */}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-gradient-to-b from-background via-background to-muted/20"
			/>
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:4rem_4rem] opacity-[0.04] [mask-image:radial-gradient(ellipse_80%_50%_at_50%_0%,black_20%,transparent_100%)]"
			/>
			<div
				aria-hidden="true"
				className="pointer-events-none absolute -top-32 left-1/2 h-[400px] w-[600px] -translate-x-1/2 rounded-full bg-primary/[0.03] blur-[120px]"
			/>

			<div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* ── Mini CTA strip ── */}
				<div className="flex flex-col items-center justify-between gap-4 border-b border-border/40 py-10 sm:flex-row sm:py-12">
					<div className="text-center sm:text-left">
						<h3 className="text-lg font-semibold text-foreground sm:text-xl">
							Pronto para automatizar seu WhatsApp?
						</h3>
						<p className="mt-1.5 text-sm text-muted-foreground">
							Comece grátis em 5 minutos. Sem cartão, sem
							aprovação manual.
						</p>
					</div>
					<div className="flex items-center gap-3">
						<Button
							size="sm"
							asChild
							className="h-9 px-5 text-sm font-medium shadow-lg shadow-primary/20"
						>
							<Link href="/sign-up">
								Criar Conta
								<ArrowRightIcon className="ml-1.5 size-3.5" />
							</Link>
						</Button>
						<Button
							variant="outline"
							size="sm"
							asChild
							className="h-9 px-5 text-sm font-medium"
						>
							<a
								href="https://api.zedaapi.com/docs"
								target="_blank"
								rel="noopener noreferrer"
							>
								Ver Docs
							</a>
						</Button>
					</div>
				</div>

				{/* ── Main link grid ── */}
				<div className="grid grid-cols-2 gap-x-8 gap-y-10 py-12 sm:py-14 md:grid-cols-6">
					{/* Brand column */}
					<div className="col-span-2">
						<Link
							href="/"
							className="group inline-flex items-center gap-2.5 rounded-lg transition-all duration-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
						>
							<div className="relative">
								<div
									aria-hidden="true"
									className="pointer-events-none absolute inset-0 rounded-full bg-primary/20 blur-md opacity-0 transition-opacity duration-300 group-hover:opacity-100"
								/>
								<Image
									src="/favicon-96x96.png"
									alt="Zé da API"
									width={36}
									height={36}
									className="relative rounded-full"
								/>
							</div>
							<span className="text-base font-semibold tracking-tight text-foreground">
								Zé da API
							</span>
						</Link>

						<p className="mt-4 max-w-xs text-sm leading-relaxed text-muted-foreground">
							A API de WhatsApp mais confiável do Brasil. Envie
							mensagens, gerencie instâncias e automatize sua
							comunicação com preço fixo.
						</p>

						{/* Status indicator */}
						<div className="mt-5 flex items-center gap-2">
							<span className="relative flex size-2">
								<span className="absolute inline-flex size-full animate-ping rounded-full bg-emerald-400/60" />
								<span className="relative inline-flex size-2 rounded-full bg-emerald-500" />
							</span>
							<span className="text-xs font-medium text-emerald-500">
								Todos os sistemas operacionais
							</span>
						</div>

						{/* Social links */}
						<div className="mt-5 flex items-center gap-1.5">
							{socialLinks.map((social) => (
								<a
									key={social.label}
									href={social.href}
									target="_blank"
									rel="noopener noreferrer"
									className="group/social flex size-9 items-center justify-center rounded-lg border border-transparent text-muted-foreground transition-all duration-200 hover:border-border/60 hover:bg-muted/50 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
									aria-label={social.label}
								>
									<social.icon className="size-4 transition-transform duration-200 group-hover/social:scale-110" />
								</a>
							))}
						</div>
					</div>

					{/* Produto */}
					<div>
						<h3 className="text-xs font-semibold uppercase tracking-wider text-foreground/70">
							Produto
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{productLinks.map((link) => (
								<li key={`${link.href}-${link.label}`}>
									<FooterLinkItem link={link} />
								</li>
							))}
						</ul>
					</div>

					{/* Recursos */}
					<div>
						<h3 className="text-xs font-semibold uppercase tracking-wider text-foreground/70">
							Recursos
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{resourceLinks.map((link) => (
								<li key={`${link.href}-${link.label}`}>
									<FooterLinkItem link={link} />
								</li>
							))}
						</ul>
					</div>

					{/* Legal */}
					<div>
						<h3 className="text-xs font-semibold uppercase tracking-wider text-foreground/70">
							Legal
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{legalLinks.map((link) => (
								<li key={link.href}>
									<FooterLinkItem link={link} />
								</li>
							))}
						</ul>
					</div>

					{/* Contato */}
					<div>
						<h3 className="text-xs font-semibold uppercase tracking-wider text-foreground/70">
							Contato
						</h3>
						<ul className="mt-4 flex flex-col gap-3">
							{contactInfo.map((item) => {
								const isExternal = item.href.startsWith("http");
								return (
									<li key={item.href}>
										<a
											href={item.href}
											target={
												isExternal
													? "_blank"
													: undefined
											}
											rel={
												isExternal
													? "noopener noreferrer"
													: undefined
											}
											className="group/contact inline-flex items-center gap-2.5 rounded-sm text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
										>
											<span className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted/60 transition-colors duration-200 group-hover/contact:bg-muted">
												<item.icon className="size-3.5" />
											</span>
											<span className="text-xs sm:text-sm">
												{item.label}
											</span>
										</a>
									</li>
								);
							})}
						</ul>

						{/* Location */}
						<div className="mt-5 flex items-center gap-2 text-muted-foreground/60">
							<MapPinIcon className="size-3.5 shrink-0" />
							<span className="text-xs">
								Rio de Janeiro, RJ - Brasil
							</span>
						</div>
					</div>
				</div>

				{/* ── Bottom bar ── */}
				<div className="flex flex-col items-center gap-4 border-t border-border/40 py-6 sm:flex-row sm:justify-between">
					{/* Left: Company info */}
					<div className="flex flex-col items-center gap-2 text-center sm:flex-row sm:gap-3 sm:text-left">
						<p className="text-xs text-muted-foreground">
							&copy; {new Date().getFullYear()} Setup Automatizado
							Ltda
						</p>
						<span
							aria-hidden="true"
							className="hidden size-0.5 rounded-full bg-muted-foreground/30 sm:block"
						/>
						<p className="text-xs text-muted-foreground/60">
							CNPJ: 54.246.473/0001-00
						</p>
						<span
							aria-hidden="true"
							className="hidden size-0.5 rounded-full bg-muted-foreground/30 sm:block"
						/>
						<div className="flex items-center gap-1.5 text-xs text-muted-foreground/60">
							<ShieldCheckIcon className="size-3" />
							<span>Conforme LGPD</span>
						</div>
					</div>

					{/* Right: Cookie preferences */}
					<button
						type="button"
						onClick={() =>
							window.dispatchEvent(
								new CustomEvent("open-cookie-preferences"),
							)
						}
						className="rounded-sm text-xs text-muted-foreground/60 underline-offset-4 transition-colors duration-200 hover:text-foreground hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
					>
						Preferências de cookies
					</button>
				</div>
			</div>
		</footer>
	);
}
