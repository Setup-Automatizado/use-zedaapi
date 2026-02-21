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

interface PixChargeCreatedEmailProps {
	userName?: string;
	amount?: string;
	pixCode?: string;
	expiresAt?: string;
	invoiceId?: string;
	dashboardUrl?: string;
}

export default function PixChargeCreatedEmail({
	userName = "Usuário",
	amount = "R$ 197,00",
	pixCode = "00020126580014br.gov.bcb.pix0136abc123-def456-ghi789",
	expiresAt,
	invoiceId = "INV-2026-0042",
	dashboardUrl = "https://zedaapi.com/faturamento",
}: PixChargeCreatedEmailProps) {
	const formattedExpiry = expiresAt
		? new Date(expiresAt).toLocaleDateString("pt-BR", {
				day: "2-digit",
				month: "2-digit",
				year: "numeric",
				hour: "2-digit",
				minute: "2-digit",
			})
		: "—";

	return (
		<Html>
			<Head />
			<Preview>PIX gerado — {amount} — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Pagamento via PIX</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Sua cobrança PIX foi gerada. Copie o código abaixo ou
						acesse o painel para visualizar o QR Code.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Fatura:</strong> {invoiceId}
						</Text>
						<Text style={detailRow}>
							<strong>Valor:</strong> {amount}
						</Text>
						<Text style={detailRow}>
							<strong>Vencimento:</strong> {formattedExpiry}
						</Text>
					</Section>

					<Section style={pixCodeBox}>
						<Text style={pixCodeLabel}>PIX Copia e Cola:</Text>
						<Text style={pixCodeText}>{pixCode}</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={dashboardUrl} style={ctaButton}>
							Ver QR Code
						</Link>
					</Section>

					<Text style={noteText}>
						O PIX é processado instantaneamente. Após o pagamento,
						sua assinatura será ativada automaticamente.
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

const pixCodeBox = {
	backgroundColor: "#ecfdf5",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "20px 0",
	border: "1px solid #a7f3d0",
};

const pixCodeLabel = {
	color: "#065f46",
	fontSize: "13px",
	fontWeight: "600" as const,
	margin: "0 0 8px",
};

const pixCodeText = {
	color: "#065f46",
	fontSize: "12px",
	lineHeight: "18px",
	margin: "0",
	wordBreak: "break-all" as const,
	fontFamily: "monospace",
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

const noteText = {
	color: "#64748b",
	fontSize: "13px",
	lineHeight: "22px",
	margin: "8px 0",
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
