import {
	Body,
	Container,
	Head,
	Heading,
	Hr,
	Html,
	Preview,
	Section,
	Text,
} from "@react-email/components";

interface ApiKeyRevokedEmailProps {
	userName?: string;
	keyName?: string;
	keyPrefix?: string;
	revokedAt?: string;
}

export default function ApiKeyRevokedEmail({
	userName = "Usuário",
	keyName = "Produção",
	keyPrefix = "zk_live_abc1...xyz9",
	revokedAt,
}: ApiKeyRevokedEmailProps) {
	const formattedDate = revokedAt
		? new Date(revokedAt).toLocaleDateString("pt-BR", {
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
			<Preview>Chave API revogada — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Chave API Revogada</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Uma chave API foi revogada na sua conta. Ela não poderá
						mais ser usada para autenticação.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Nome:</strong> {keyName}
						</Text>
						<Text style={detailRow}>
							<strong>Chave:</strong> {keyPrefix}
						</Text>
						<Text style={detailRow}>
							<strong>Revogada em:</strong> {formattedDate}
						</Text>
					</Section>

					<Text style={noteText}>
						Atualize suas integrações que utilizam esta chave para
						evitar interrupções no serviço.
					</Text>

					<Text style={warningText}>
						Se você não revogou esta chave, entre em contato com
						nosso suporte imediatamente.
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

const noteText = {
	color: "#64748b",
	fontSize: "13px",
	lineHeight: "22px",
	margin: "8px 0",
	textAlign: "center" as const,
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
