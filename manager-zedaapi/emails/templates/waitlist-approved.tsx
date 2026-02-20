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

interface WaitlistApprovedEmailProps {
	userName?: string;
	signUpUrl?: string;
}

export default function WaitlistApprovedEmail({
	userName = "Usuário",
	signUpUrl = "https://manager.zedaapi.com/sign-up",
}: WaitlistApprovedEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Sua conta foi aprovada! — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Sua conta foi aprovada!</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Temos boas notícias! Sua solicitação de acesso ao Zé da
						API Manager foi aprovada. Agora você pode criar sua
						conta e começar a gerenciar suas instâncias WhatsApp.
					</Text>

					<Section style={ctaSection}>
						<Link href={signUpUrl} style={ctaButton}>
							Criar Minha Conta
						</Link>
					</Section>

					<Text style={paragraph}>
						Com o Zé da API Manager você pode:
					</Text>
					<Section style={featureList}>
						<Text style={feature}>
							Gerenciar múltiplas instâncias WhatsApp
						</Text>
						<Text style={feature}>
							Monitorar conexões em tempo real
						</Text>
						<Text style={feature}>
							Configurar webhooks personalizados
						</Text>
						<Text style={feature}>
							Emitir NFS-e automaticamente
						</Text>
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
	color: "#16a34a",
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

const ctaSection = {
	textAlign: "center" as const,
	margin: "28px 0",
};

const ctaButton = {
	backgroundColor: "#16a34a",
	borderRadius: "6px",
	color: "#ffffff",
	fontSize: "15px",
	fontWeight: "600" as const,
	textDecoration: "none",
	padding: "14px 36px",
	display: "inline-block" as const,
};

const featureList = {
	backgroundColor: "#f0fdf4",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "16px 0",
	border: "1px solid #bbf7d0",
};

const feature = {
	color: "#15803d",
	fontSize: "14px",
	lineHeight: "26px",
	margin: "0",
	paddingLeft: "8px",
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
