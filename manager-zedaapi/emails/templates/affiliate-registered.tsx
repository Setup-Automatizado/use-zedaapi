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

interface AffiliateRegisteredEmailProps {
	userName?: string;
	affiliateCode?: string;
	commissionRate?: string;
	dashboardUrl?: string;
}

export default function AffiliateRegisteredEmail({
	userName = "Usuário",
	affiliateCode = "REF-ABC123",
	commissionRate = "20%",
	dashboardUrl = "https://zedaapi.com/afiliados",
}: AffiliateRegisteredEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Cadastro de afiliado confirmado — Zé da API</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>
						Bem-vindo ao Programa de Afiliados!
					</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Seu cadastro no programa de afiliados foi aprovado.
						Agora você pode indicar novos clientes e ganhar
						comissões.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Código de indicação:</strong>{" "}
							{affiliateCode}
						</Text>
						<Text style={detailRow}>
							<strong>Comissão:</strong> {commissionRate} por
							venda
						</Text>
					</Section>

					<Section style={steps}>
						<Text style={stepTitle}>Como funciona:</Text>
						<Text style={step}>
							1. Compartilhe seu link de indicação
						</Text>
						<Text style={step}>
							2. Novos clientes se cadastram pelo seu link
						</Text>
						<Text style={step}>
							3. Quando assinam, você ganha comissão
						</Text>
						<Text style={step}>
							4. Solicite o saque quando quiser
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={dashboardUrl} style={ctaButton}>
							Acessar Painel de Afiliados
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
	fontSize: "22px",
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

const detailsBox = {
	backgroundColor: "#f8fafc",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #e2e8f0",
};

const detailRow = {
	color: "#334155",
	fontSize: "14px",
	lineHeight: "28px",
	margin: "0",
};

const steps = {
	backgroundColor: "#f0fdf4",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #bbf7d0",
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
