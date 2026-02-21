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

interface BoletoChargeCreatedEmailProps {
	userName?: string;
	amount?: string;
	linhaDigitavel?: string;
	dueDate?: string;
	invoiceId?: string;
	boletoUrl?: string;
}

export default function BoletoChargeCreatedEmail({
	userName = "Usuário",
	amount = "R$ 197,00",
	linhaDigitavel = "23793.38128 60000.000003 00000.000400 1 84340000019700",
	dueDate,
	invoiceId = "INV-2026-0042",
	boletoUrl = "https://zedaapi.com/faturamento",
}: BoletoChargeCreatedEmailProps) {
	const formattedDate = dueDate
		? new Date(dueDate).toLocaleDateString("pt-BR")
		: "—";

	return (
		<Html>
			<Head />
			<Preview>Boleto gerado — {amount} — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Pagamento via Boleto</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Seu boleto foi gerado. Copie a linha digitável abaixo ou
						clique no botão para visualizar o boleto completo.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Fatura:</strong> {invoiceId}
						</Text>
						<Text style={detailRow}>
							<strong>Valor:</strong> {amount}
						</Text>
						<Text style={detailRow}>
							<strong>Vencimento:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={barcodeBox}>
						<Text style={barcodeLabel}>Linha Digitável:</Text>
						<Text style={barcodeText}>{linhaDigitavel}</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={boletoUrl} style={ctaButton}>
							Ver Boleto Completo
						</Link>
					</Section>

					<Text style={noteText}>
						O boleto pode levar até 3 dias úteis para ser
						compensado. Após a confirmação, sua assinatura será
						ativada automaticamente.
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

const barcodeBox = {
	backgroundColor: "#f0f9ff",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "20px 0",
	border: "1px solid #bae6fd",
};

const barcodeLabel = {
	color: "#0c4a6e",
	fontSize: "13px",
	fontWeight: "600" as const,
	margin: "0 0 8px",
};

const barcodeText = {
	color: "#0c4a6e",
	fontSize: "14px",
	lineHeight: "22px",
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
