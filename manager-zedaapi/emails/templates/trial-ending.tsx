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

interface TrialEndingEmailProps {
	userName?: string;
	planName?: string;
	daysRemaining?: number;
	trialEndDate?: string;
	upgradeUrl?: string;
}

export default function TrialEndingEmail({
	userName = "Usuário",
	planName = "Profissional",
	daysRemaining = 3,
	trialEndDate,
	upgradeUrl = "https://zedaapi.com/assinaturas",
}: TrialEndingEmailProps) {
	const formattedEndDate = trialEndDate
		? new Date(trialEndDate).toLocaleDateString("pt-BR")
		: "—";

	return (
		<Html>
			<Head />
			<Preview>
				{`Seu teste encerra em ${daysRemaining} dias — Zé da API Manager`}
			</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>
						Seu Teste Encerra em {daysRemaining} Dias
					</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Seu período de teste do plano {planName} está chegando
						ao fim. Para continuar usando todos os recursos sem
						interrupção, assine agora.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Plano:</strong> {planName}
						</Text>
						<Text style={detailRow}>
							<strong>Dias restantes:</strong> {daysRemaining}
						</Text>
						<Text style={detailRow}>
							<strong>Encerra em:</strong> {formattedEndDate}
						</Text>
					</Section>

					<Section style={warningBox}>
						<Text style={warningText}>
							Após o término do teste, suas instâncias serão
							desconectadas e o acesso aos recursos premium será
							suspenso.
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={upgradeUrl} style={ctaButton}>
							Assinar Agora
						</Link>
					</Section>

					<Hr style={hr} />
					<Text style={footer}>
						Dúvidas? Entre em contato com nosso suporte.
					</Text>
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
	color: "#d97706",
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
	backgroundColor: "#fffbeb",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #fde68a",
};

const detailRow = {
	color: "#334155",
	fontSize: "14px",
	lineHeight: "28px",
	margin: "0",
};

const warningBox = {
	backgroundColor: "#fef2f2",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "16px 0",
	border: "1px solid #fecaca",
};

const warningText = {
	color: "#b91c1c",
	fontSize: "14px",
	lineHeight: "22px",
	margin: "0",
};

const ctaSection = {
	textAlign: "center" as const,
	margin: "24px 0",
};

const ctaButton = {
	backgroundColor: "#d97706",
	borderRadius: "6px",
	color: "#ffffff",
	fontSize: "15px",
	fontWeight: "600" as const,
	textDecoration: "none",
	padding: "14px 36px",
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
	margin: "4px 0",
};
