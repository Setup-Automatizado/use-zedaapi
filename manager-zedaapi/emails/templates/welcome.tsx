import {
	Body,
	Container,
	Head,
	Heading,
	Hr,
	Html,
	Link,
	Preview,
	Section,
	Text,
} from "@react-email/components";

interface WelcomeEmailProps {
	userName?: string;
	dashboardUrl?: string;
}

export default function WelcomeEmail({
	userName = "Usuário",
	dashboardUrl = "https://zedaapi.com/painel",
}: WelcomeEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Bem-vindo ao Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>
						Bem-vindo ao Zé da API Manager!
					</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Sua conta foi criada com sucesso. Agora você pode
						gerenciar suas instâncias WhatsApp de forma simples e
						profissional.
					</Text>

					<Section style={steps}>
						<Text style={stepTitle}>Primeiros passos:</Text>
						<Text style={step}>
							1. Escolha um plano de assinatura
						</Text>
						<Text style={step}>
							2. Crie sua primeira instância WhatsApp
						</Text>
						<Text style={step}>
							3. Conecte escaneando o QR Code
						</Text>
						<Text style={step}>4. Configure seus webhooks</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={dashboardUrl} style={ctaButton}>
							Acessar Painel
						</Link>
					</Section>

					<Hr style={hr} />
					<Text style={footer}>
						Zé da API Manager — Gerencie suas instâncias WhatsApp
					</Text>
				</Container>
			</Body>
		</Html>
	);
}

const main = {
	backgroundColor: "#f6f9fc",
	fontFamily:
		'-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Ubuntu, sans-serif',
};

const container = {
	backgroundColor: "#ffffff",
	margin: "0 auto",
	padding: "40px 20px",
	maxWidth: "560px",
	borderRadius: "8px",
};

const logo = {
	color: "#0f172a",
	fontSize: "20px",
	fontWeight: "700" as const,
	textAlign: "center" as const,
	margin: "0 0 20px",
};

const heading = {
	color: "#0f172a",
	fontSize: "24px",
	fontWeight: "600" as const,
	textAlign: "center" as const,
	margin: "20px 0",
};

const paragraph = {
	color: "#334155",
	fontSize: "15px",
	lineHeight: "26px",
	margin: "8px 0",
};

const steps = {
	backgroundColor: "#f8fafc",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
};

const stepTitle = {
	color: "#0f172a",
	fontSize: "14px",
	fontWeight: "600" as const,
	margin: "0 0 12px",
};

const step = {
	color: "#475569",
	fontSize: "14px",
	lineHeight: "24px",
	margin: "4px 0",
};

const ctaSection = {
	textAlign: "center" as const,
	margin: "24px 0",
};

const ctaButton = {
	backgroundColor: "#0f172a",
	borderRadius: "6px",
	color: "#ffffff",
	fontSize: "14px",
	fontWeight: "600" as const,
	textDecoration: "none",
	padding: "12px 32px",
	display: "inline-block" as const,
};

const hr = {
	borderColor: "#e2e8f0",
	margin: "20px 0",
};

const footer = {
	color: "#94a3b8",
	fontSize: "12px",
	textAlign: "center" as const,
	margin: "20px 0 0",
};
