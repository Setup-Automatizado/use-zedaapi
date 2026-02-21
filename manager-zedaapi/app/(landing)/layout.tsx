import type { Metadata } from "next";
import { Header } from "@/components/landing/header";
import { Footer } from "@/components/landing/footer";
import { WhatsAppWidget } from "@/components/landing/whatsapp-widget";
import { CookieConsent } from "@/components/landing/cookie-consent";
import { ScrollToTop } from "@/components/shared/scroll-to-top";

export const metadata: Metadata = {
	title: "Zé da API - WhatsApp API para Desenvolvedores",
	description:
		"API profissional para WhatsApp. Envie mensagens, gerencie instâncias e integre com seus sistemas usando a API mais confiável do Brasil.",
	openGraph: {
		title: "Zé da API - WhatsApp API para Desenvolvedores",
		description:
			"API profissional para WhatsApp. Envie mensagens, gerencie instâncias e integre com seus sistemas usando a API mais confiável do Brasil.",
		url: "https://zedaapi.com",
		siteName: "Zé da API",
		locale: "pt_BR",
		type: "website",
	},
	twitter: {
		card: "summary_large_image",
		title: "Zé da API - WhatsApp API para Desenvolvedores",
		description:
			"API profissional para WhatsApp. Envie mensagens, gerencie instâncias e integre com seus sistemas.",
	},
};

export default function LandingLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<div className="flex min-h-svh flex-col">
			<Header />
			<main className="flex-1">{children}</main>
			<Footer />
			<ScrollToTop />
			<WhatsAppWidget />
			<CookieConsent />
		</div>
	);
}
