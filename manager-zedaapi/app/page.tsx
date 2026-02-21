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
import { ScrollToTop } from "@/components/shared/scroll-to-top";

export const metadata: Metadata = {
	title: "Zé da API - API de WhatsApp para Empresas | Envie Mensagens via API REST",
	description:
		"Automatize o WhatsApp da sua empresa com a ZedaAPI. Até 10M mensagens/mês, webhooks em tempo real, 99.9% de uptime e preço fixo sem taxa por mensagem. +500 empresas confiam. Teste grátis por 7 dias, sem cartão.",
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
		"API WhatsApp preço",
		"sem taxa por mensagem",
	],
	openGraph: {
		title: "Zé da API - API de WhatsApp para Empresas",
		description:
			"Envie até 10M mensagens/mês via API REST com 47ms de latência. Webhooks em tempo real, multi-instância e preço fixo. +500 empresas confiam. Teste grátis, sem cartão.",
		url: "https://zedaapi.com",
		siteName: "Zé da API",
		locale: "pt_BR",
		type: "website",
	},
	twitter: {
		card: "summary_large_image",
		title: "Zé da API - API de WhatsApp para Empresas",
		description:
			"Automatize o WhatsApp da sua empresa. API REST completa, webhooks em tempo real, multi-device. Teste grátis por 7 dias, sem cartão.",
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

const organizationJsonLd = {
	"@context": "https://schema.org",
	"@type": "Organization",
	name: "Zé da API",
	alternateName: "ZedaAPI",
	url: "https://zedaapi.com",
	logo: "https://zedaapi.com/logo.png",
	description:
		"API profissional de WhatsApp para empresas e desenvolvedores brasileiros.",
	contactPoint: [
		{
			"@type": "ContactPoint",
			telephone: "+55-21-97153-2700",
			contactType: "customer service",
			availableLanguage: "Portuguese",
			areaServed: "BR",
		},
		{
			"@type": "ContactPoint",
			email: "suporte@zedaapi.com",
			contactType: "technical support",
			availableLanguage: "Portuguese",
		},
	],
	address: {
		"@type": "PostalAddress",
		addressLocality: "Rio de Janeiro",
		addressRegion: "RJ",
		addressCountry: "BR",
	},
	sameAs: [],
};

const softwareJsonLd = {
	"@context": "https://schema.org",
	"@type": "SoftwareApplication",
	name: "Zé da API - WhatsApp API",
	applicationCategory: "BusinessApplication",
	operatingSystem: "Web",
	description:
		"API REST completa para integracao com WhatsApp. Envie mensagens, gerencie instancias e automatize comunicacao.",
	url: "https://zedaapi.com",
	offers: {
		"@type": "AggregateOffer",
		priceCurrency: "BRL",
		lowPrice: "97",
		highPrice: "997",
		offerCount: "4",
	},
	featureList: [
		"API REST para WhatsApp",
		"Webhooks em tempo real",
		"Multi-instancia",
		"Envio de texto, imagem, video, audio, documentos",
		"99.9% uptime SLA",
		"Preco fixo sem taxa por mensagem",
	],
};

export default function LandingPage() {
	return (
		<div className="flex min-h-svh flex-col">
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify(organizationJsonLd),
				}}
			/>
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{
					__html: JSON.stringify(softwareJsonLd),
				}}
			/>
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
			<ScrollToTop />
			<WhatsAppWidget />
			<CookieConsent />
		</div>
	);
}
