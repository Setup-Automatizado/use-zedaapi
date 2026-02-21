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

interface SubscriptionRenewedEmailProps {
	userName?: string;
	planName?: string;
	amount?: string;
	nextRenewalDate?: string;
	dashboardUrl?: string;
}

export default function SubscriptionRenewedEmail({
	userName = "Usuário",
	planName = "Profissional",
	amount = "R$ 197,00",
	nextRenewalDate,
	dashboardUrl = "https://zedaapi.com/assinaturas",
}: SubscriptionRenewedEmailProps) {
	const formattedDate = nextRenewalDate
		? new Date(nextRenewalDate).toLocaleDateString("pt-BR")
		: "—";

	return (
		<Html>
			<Head />
			<Preview>Assinatura renovada — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Assinatura Renovada</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Sua assinatura foi renovada automaticamente com sucesso.
						Seus serviços continuam funcionando normalmente.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Plano:</strong> {planName}
						</Text>
						<Text style={detailRow}>
							<strong>Valor cobrado:</strong> {amount}
						</Text>
						<Text style={detailRow}>
							<strong>Próxima renovação:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={dashboardUrl} style={ctaButton}>
							Ver Assinatura
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
	color: "#16a34a",
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
