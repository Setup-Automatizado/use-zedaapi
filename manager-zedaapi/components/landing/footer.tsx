"use client";

import Link from "next/link";
import {
	MailIcon,
	PhoneIcon,
	HeadphonesIcon,
	GithubIcon,
	LinkedinIcon,
	TwitterIcon,
} from "lucide-react";

const productLinks = [
	{ label: "Recursos", href: "#recursos" },
	{ label: "Preços", href: "#precos" },
	{ label: "Integrações", href: "#integracoes" },
	{ label: "Documentação", href: "/docs" },
	{ label: "API Reference", href: "/docs/api" },
	{ label: "Status", href: "/status" },
] as const;

const resourceLinks = [
	{ label: "Blog", href: "/blog" },
	{ label: "Changelog", href: "/changelog" },
	{ label: "Guias de Integração", href: "/docs/guides" },
	{ label: "Postman Collection", href: "/docs/postman" },
	{ label: "Suporte", href: "/suporte" },
] as const;

const legalLinks = [
	{ label: "Termos de Uso", href: "/termos-de-uso" },
	{ label: "Política de Privacidade", href: "/politica-de-privacidade" },
	{ label: "Política de Cookies", href: "/politica-de-cookies" },
	{ label: "LGPD", href: "/lgpd" },
	{ label: "Exclusão de Dados", href: "/exclusao-de-dados" },
] as const;

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

export function Footer() {
	return (
		<footer className="border-t border-border/50 bg-muted/10">
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Main grid */}
				<div className="grid grid-cols-2 gap-8 py-14 sm:py-16 md:grid-cols-5">
					{/* Brand column */}
					<div className="col-span-2 md:col-span-1">
						<Link
							href="/"
							className="flex items-center gap-2.5 transition-opacity duration-200 hover:opacity-80"
						>
							<div className="flex size-8 items-center justify-center rounded-xl bg-primary font-bold text-sm text-primary-foreground shadow-sm">
								Z
							</div>
							<span className="text-base font-semibold tracking-tight text-foreground">
								Zé da API
							</span>
						</Link>
						<p className="mt-4 text-sm leading-relaxed text-muted-foreground">
							A API de WhatsApp mais confiável do Brasil. Envie
							mensagens, gerencie instâncias e automatize sua
							comunicação.
						</p>

						{/* Social links */}
						<div className="mt-5 flex items-center gap-2">
							{socialLinks.map((social) => (
								<a
									key={social.label}
									href={social.href}
									target="_blank"
									rel="noopener noreferrer"
									className="flex size-8 items-center justify-center rounded-lg text-muted-foreground transition-colors duration-200 hover:bg-muted hover:text-foreground"
									aria-label={social.label}
								>
									<social.icon className="size-4" />
								</a>
							))}
						</div>
					</div>

					{/* Produto */}
					<div>
						<h3 className="text-sm font-semibold text-foreground">
							Produto
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{productLinks.map((link) => (
								<li key={link.href}>
									{link.href.startsWith("#") ? (
										<a
											href={link.href}
											className="text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
										>
											{link.label}
										</a>
									) : (
										<Link
											href={link.href}
											className="text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
										>
											{link.label}
										</Link>
									)}
								</li>
							))}
						</ul>
					</div>

					{/* Recursos */}
					<div>
						<h3 className="text-sm font-semibold text-foreground">
							Recursos
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{resourceLinks.map((link) => (
								<li key={link.href}>
									<Link
										href={link.href}
										className="text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
									>
										{link.label}
									</Link>
								</li>
							))}
						</ul>
					</div>

					{/* Legal */}
					<div>
						<h3 className="text-sm font-semibold text-foreground">
							Legal
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{legalLinks.map((link) => (
								<li key={link.href}>
									<Link
										href={link.href}
										className="text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
									>
										{link.label}
									</Link>
								</li>
							))}
						</ul>
					</div>

					{/* Contato */}
					<div>
						<h3 className="text-sm font-semibold text-foreground">
							Contato
						</h3>
						<ul className="mt-4 flex flex-col gap-2.5">
							{contactInfo.map((item) => (
								<li key={item.href}>
									<a
										href={item.href}
										target={
											item.href.startsWith("http")
												? "_blank"
												: undefined
										}
										rel={
											item.href.startsWith("http")
												? "noopener noreferrer"
												: undefined
										}
										className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
									>
										<item.icon className="size-3.5 shrink-0" />
										{item.label}
									</a>
								</li>
							))}
						</ul>
						<p className="mt-4 text-xs leading-relaxed text-muted-foreground/70">
							Rio de Janeiro, RJ - Brasil
						</p>
					</div>
				</div>

				{/* Bottom bar */}
				<div className="flex flex-col items-center gap-4 border-t border-border/50 py-6 text-center sm:flex-row sm:justify-between sm:text-left">
					<div className="flex flex-col gap-1">
						<p className="text-xs text-muted-foreground">
							&copy; {new Date().getFullYear()} Setup Automatizado
							Ltda
						</p>
						<p className="text-xs text-muted-foreground/70">
							CNPJ: 54.246.473/0001-00 &middot; Rio de Janeiro, RJ
							- Brasil
						</p>
					</div>
					<button
						type="button"
						onClick={() =>
							window.dispatchEvent(
								new CustomEvent("open-cookie-preferences"),
							)
						}
						className="text-xs text-muted-foreground underline-offset-4 transition-colors duration-200 hover:text-foreground hover:underline"
					>
						Alterar preferências de cookies
					</button>
				</div>
			</div>
		</footer>
	);
}
