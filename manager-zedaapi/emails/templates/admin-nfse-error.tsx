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

interface AdminNfseErrorEmailProps {
	adminName?: string;
	userName?: string;
	userEmail?: string;
	invoiceId?: string;
	errorMessage?: string;
	errorCode?: string;
	failedAt?: string;
	adminUrl?: string;
}

export default function AdminNfseErrorEmail({
	adminName = "Administrador",
	userName = "João Silva",
	userEmail = "joao@exemplo.com",
	invoiceId = "INV-2026-0042",
	errorMessage = "Erro ao comunicar com a prefeitura: timeout",
	errorCode = "NFSE_TIMEOUT",
	failedAt,
	adminUrl = "https://zedaapi.com/admin/nfe",
}: AdminNfseErrorEmailProps) {
	const formattedDate = failedAt
		? new Date(failedAt).toLocaleDateString("pt-BR", {
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
			<Preview>Erro NFS-e: {errorCode} — Admin Zé da API</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Erro na Emissão de NFS-e</Heading>
					<Text style={paragraph}>Olá {adminName},</Text>
					<Text style={paragraph}>
						Ocorreu um erro ao emitir a NFS-e. Ação manual pode ser
						necessária.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Cliente:</strong> {userName} ({userEmail})
						</Text>
						<Text style={detailRow}>
							<strong>Fatura:</strong> {invoiceId}
						</Text>
						<Text style={detailRow}>
							<strong>Código:</strong> {errorCode}
						</Text>
						<Text style={detailRow}>
							<strong>Erro:</strong> {errorMessage}
						</Text>
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={adminUrl} style={ctaButton}>
							Ver no Admin
						</Link>
					</Section>

					<Hr style={hr} />
					<Text style={footer}>
						Notificação administrativa — Zé da API Manager
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
