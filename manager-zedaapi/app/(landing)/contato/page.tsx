import type { Metadata } from "next";
import { Suspense } from "react";
import {
	MailIcon,
	MessageSquareIcon,
	MapPinIcon,
	ClockIcon,
	HeadphonesIcon,
} from "lucide-react";
import { ContactForm } from "@/components/landing/contact-form";

export const metadata: Metadata = {
	title: "Contato - Ze da API",
	description:
		"Entre em contato com a equipe da Ze da API. Estamos prontos para ajudar com duvidas comerciais, suporte tecnico e parcerias.",
	openGraph: {
		title: "Contato - Ze da API",
		description:
			"Entre em contato com a equipe da Ze da API. Suporte comercial e tecnico.",
		url: "https://zedaapi.com/contato",
	},
	alternates: {
		canonical: "https://zedaapi.com/contato",
	},
};

const contactChannels = [
	{
		icon: MailIcon,
		title: "E-mail",
		value: "contato@zedaapi.com",
		href: "mailto:contato@zedaapi.com",
		description: "Respondemos em ate 24h uteis",
	},
	{
		icon: MessageSquareIcon,
		title: "WhatsApp",
		value: "+55 21 97153-2700",
		href: "https://wa.me/5521971532700",
		description: "Atendimento em horario comercial",
	},
	{
		icon: MapPinIcon,
		title: "Endereco",
		value: "Rio de Janeiro, RJ - Brasil",
		href: null,
		description: "Atendemos 100% remoto",
	},
	{
		icon: ClockIcon,
		title: "Horario de atendimento",
		value: "Seg a Sex, 9h as 18h",
		href: null,
		description: "Fuso horario de Brasilia (BRT)",
	},
] as const;

export default function ContatoPage() {
	return (
		<div className="relative">
			{/* Hero */}
			<section className="border-b border-border bg-muted/30 py-16 sm:py-20">
				<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<div className="mx-auto max-w-2xl text-center">
						<h1 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
							Entre em Contato
						</h1>
						<p className="mt-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
							Tem dúvidas, quer uma demonstração ou precisa de
							suporte? Nossa equipe está pronta para ajudar.
						</p>
					</div>
				</div>
			</section>

			{/* Content */}
			<section className="py-12 sm:py-16 lg:py-20">
				<div className="mx-auto grid max-w-7xl grid-cols-1 gap-12 px-4 sm:px-6 lg:grid-cols-5 lg:gap-16 lg:px-8">
					{/* Left: Contact info (2 cols) */}
					<div className="lg:col-span-2">
						<h2 className="text-xl font-semibold text-foreground">
							Canais de atendimento
						</h2>
						<p className="mt-2 text-sm text-muted-foreground">
							Escolha o canal mais conveniente para você.
						</p>

						<div className="mt-8 flex flex-col gap-6">
							{contactChannels.map((channel) => {
								const Icon = channel.icon;
								return (
									<div
										key={channel.title}
										className="flex gap-4"
									>
										<div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
											<Icon className="size-5" />
										</div>
										<div>
											<p className="text-sm font-medium text-foreground">
												{channel.title}
											</p>
											{channel.href ? (
												<a
													href={channel.href}
													target={
														channel.href.startsWith(
															"http",
														)
															? "_blank"
															: undefined
													}
													rel={
														channel.href.startsWith(
															"http",
														)
															? "noopener noreferrer"
															: undefined
													}
													className="text-sm text-primary underline-offset-4 hover:underline"
												>
													{channel.value}
												</a>
											) : (
												<p className="text-sm text-foreground/80">
													{channel.value}
												</p>
											)}
											<p className="mt-0.5 text-xs text-muted-foreground">
												{channel.description}
											</p>
										</div>
									</div>
								);
							})}
						</div>

						{/* Support info card */}
						<div className="mt-10 rounded-2xl border border-border bg-muted/30 p-6">
							<div className="flex items-center gap-3">
								<div className="flex size-9 items-center justify-center rounded-lg bg-primary/10">
									<HeadphonesIcon className="size-4.5 text-primary" />
								</div>
								<h3 className="text-sm font-semibold text-foreground">
									Suporte tecnico
								</h3>
							</div>
							<p className="mt-3 text-sm leading-relaxed text-muted-foreground">
								Para suporte tecnico e problemas com a API,
								envie um e-mail para{" "}
								<a
									href="mailto:suporte@zedaapi.com"
									className="font-medium text-primary underline-offset-4 hover:underline"
								>
									suporte@zedaapi.com
								</a>{" "}
								ou utilize o formulario ao lado selecionando
								&quot;Suporte Tecnico&quot; como assunto.
							</p>
						</div>
					</div>

					{/* Right: Contact form (3 cols) */}
					<div className="lg:col-span-3">
						<Suspense fallback={null}>
							<ContactForm />
						</Suspense>
					</div>
				</div>
			</section>
		</div>
	);
}
