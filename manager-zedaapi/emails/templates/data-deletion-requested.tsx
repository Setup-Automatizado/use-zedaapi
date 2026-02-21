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

interface DataDeletionRequestedEmailProps {
	userName?: string;
	requestedAt?: string;
	deletionDate?: string;
	supportUrl?: string;
}

export default function DataDeletionRequestedEmail({
	userName = "Usuário",
	requestedAt,
	deletionDate,
	supportUrl = "https://zedaapi.com/suporte",
}: DataDeletionRequestedEmailProps) {
	const formattedRequestDate = requestedAt
		? new Date(requestedAt).toLocaleDateString("pt-BR")
		: new Date().toLocaleDateString("pt-BR");

	const formattedDeletionDate = deletionDate
		? new Date(deletionDate).toLocaleDateString("pt-BR")
		: "—";

	return (
		<Html>
			<Head />
			<Preview>
				Solicitação de exclusão de dados — Zé da API Manager
			</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>
						Solicitação de Exclusão de Dados
					</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Recebemos sua solicitação de exclusão de dados pessoais.
						Seus dados serão removidos permanentemente no prazo
						indicado abaixo, conforme a LGPD.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Solicitado em:</strong>{" "}
							{formattedRequestDate}
						</Text>
						<Text style={detailRow}>
							<strong>Exclusão prevista:</strong>{" "}
							{formattedDeletionDate}
						</Text>
					</Section>

					<Text style={paragraph}>
						Após a exclusão, todos os seus dados, instâncias,
						faturas e configurações serão removidos permanentemente.
						Esta ação não pode ser desfeita.
					</Text>

					<Text style={paragraph}>
						Se você deseja cancelar esta solicitação, entre em
						contato com nosso suporte antes da data de exclusão.
					</Text>

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
