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

interface EmailChangedEmailProps {
	userName?: string;
	oldEmail?: string;
	newEmail?: string;
	changedAt?: string;
	supportUrl?: string;
}

export default function EmailChangedEmail({
	userName = "Usuário",
	oldEmail = "antigo@exemplo.com",
	newEmail = "novo@exemplo.com",
	changedAt,
	supportUrl = "https://zedaapi.com/suporte",
}: EmailChangedEmailProps) {
	const formattedDate = changedAt
		? new Date(changedAt).toLocaleDateString("pt-BR", {
				day: "2-digit",
				month: "2-digit",
				year: "numeric",
				hour: "2-digit",
				minute: "2-digit",
			})
		: new Date().toLocaleDateString("pt-BR", {
				day: "2-digit",
				month: "2-digit",
				year: "numeric",
				hour: "2-digit",
				minute: "2-digit",
			});

	return (
		<Html>
			<Head />
			<Preview>Seu e-mail foi alterado — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>E-mail Alterado</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						O e-mail associado à sua conta foi alterado com sucesso.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>E-mail anterior:</strong> {oldEmail}
						</Text>
						<Text style={detailRow}>
							<strong>Novo e-mail:</strong> {newEmail}
						</Text>
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={warningBox}>
						<Text style={warningTitle}>Não foi você?</Text>
						<Text style={warningBody}>
							Se você não realizou essa alteração, entre em
							contato com nosso suporte imediatamente. Sua conta
							pode estar comprometida.
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={supportUrl} style={ctaButton}>
							Contatar Suporte
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

const warningBox = {
	backgroundColor: "#fef2f2",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "16px 0",
	border: "1px solid #fecaca",
};

const warningTitle = {
	color: "#b91c1c",
	fontSize: "14px",
	fontWeight: "600" as const,
	margin: "0 0 8px",
};

const warningBody = {
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
	backgroundColor: "#dc2626",
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
