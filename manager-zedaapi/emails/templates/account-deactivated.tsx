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

interface AccountDeactivatedEmailProps {
	userName?: string;
	deactivatedAt?: string;
	reason?: string;
	supportUrl?: string;
}

export default function AccountDeactivatedEmail({
	userName = "Usuário",
	deactivatedAt,
	reason = "Solicitação do usuário",
	supportUrl = "https://zedaapi.com/suporte",
}: AccountDeactivatedEmailProps) {
	const formattedDate = deactivatedAt
		? new Date(deactivatedAt).toLocaleDateString("pt-BR", {
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
			<Preview>Conta desativada — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Conta Desativada</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Confirmamos que sua conta foi desativada. Todas as
						instâncias WhatsApp vinculadas foram desconectadas e
						seus dados serão mantidos por 30 dias antes da exclusão
						permanente.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
						<Text style={detailRow}>
							<strong>Motivo:</strong> {reason}
						</Text>
					</Section>

					<Text style={paragraph}>
						Se você deseja reativar sua conta dentro do período de
						30 dias, entre em contato com nosso suporte.
					</Text>

					<Section style={ctaSection}>
						<Link href={supportUrl} style={ctaButton}>
							Contatar Suporte
						</Link>
					</Section>

					<Text style={warningText}>
						Se você não solicitou a desativação, entre em contato
						com nosso suporte imediatamente.
					</Text>

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
	color: "#dc2626",
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
	backgroundColor: "#fef2f2",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #fecaca",
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

const warningText = {
	color: "#b91c1c",
	fontSize: "13px",
	lineHeight: "22px",
	margin: "16px 0",
	textAlign: "center" as const,
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
