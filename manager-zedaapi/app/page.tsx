import type { Metadata } from "next";
import { Header } from "@/components/landing/header";
import { Footer } from "@/components/landing/footer";
import { WhatsAppWidget } from "@/components/landing/whatsapp-widget";
import { CookieConsent } from "@/components/landing/cookie-consent";
import { Hero } from "@/components/landing/hero";
import { SocialProof } from "@/components/landing/social-proof";
import { Features } from "@/components/landing/features";
import { HowItWorks } from "@/components/landing/how-it-works";
import { Pricing } from "@/components/landing/pricing";
import { Integrations } from "@/components/landing/integrations";
import { FAQ } from "@/components/landing/faq";
import { CTA } from "@/components/landing/cta";

export const metadata: Metadata = {
	title: "Zé da API - A API de WhatsApp #1 do Brasil | Envie Mensagens via API REST",
	description:
		"Automatize o WhatsApp da sua empresa com a API mais confiável do Brasil. Envie mensagens, gerencie múltiplas instâncias, webhooks em tempo real e 99.9% de uptime. +500 empresas confiam. Teste grátis por 7 dias.",
	keywords: [
		"WhatsApp API",
		"API WhatsApp Brasil",
		"WhatsApp Business API",
		"automação WhatsApp",
		"enviar mensagens WhatsApp API",
		"integração WhatsApp",
		"API REST WhatsApp",
		"webhooks WhatsApp",
		"multi-device WhatsApp",
		"Zé da API",
		"ZedaAPI",
		"zedaapi.com",
	],
	openGraph: {
		title: "Zé da API - Automatize o WhatsApp da Sua Empresa",
		description:
			"A API de WhatsApp #1 do Brasil. Envie mensagens, gerencie instâncias e receba webhooks em tempo real. +500 empresas confiam. Teste grátis.",
		url: "https://zedaapi.com",
		siteName: "Zé da API",
		locale: "pt_BR",
		type: "website",
	},
	twitter: {
		card: "summary_large_image",
		title: "Zé da API - A API de WhatsApp #1 do Brasil",
		description:
			"Automatize o WhatsApp da sua empresa. API REST completa, webhooks em tempo real, multi-device. Teste grátis por 7 dias.",
	},
	alternates: {
		canonical: "https://zedaapi.com",
	},
	robots: {
		index: true,
		follow: true,
		googleBot: {
			index: true,
			follow: true,
			"max-video-preview": -1,
			"max-image-preview": "large",
			"max-snippet": -1,
		},
	},
};

export default function LandingPage() {
	return (
		<div className="flex min-h-svh flex-col">
			<Header />
			<main className="flex-1">
				<Hero />
				<SocialProof />
				<Features />
				<HowItWorks />
				<Pricing />
				<Integrations />
				<FAQ />
				<CTA />
			</main>
			<Footer />
			<WhatsAppWidget />
			<CookieConsent />
		</div>
	);
}
