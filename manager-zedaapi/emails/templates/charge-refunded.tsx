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

interface ChargeRefundedEmailProps {
	userName?: string;
	amount?: string;
	refundedAt?: string;
	invoiceId?: string;
	reason?: string;
}

export default function ChargeRefundedEmail({
	userName = "Usuário",
	amount = "R$ 197,00",
	refundedAt,
	invoiceId = "INV-2026-0042",
	reason = "Solicitação do cliente",
}: ChargeRefundedEmailProps) {
	const formattedDate = refundedAt
		? new Date(refundedAt).toLocaleDateString("pt-BR")
		: new Date().toLocaleDateString("pt-BR");

	return (
		<Html>
			<Head />
			<Preview>Reembolso processado — {amount}</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Reembolso Processado</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Seu reembolso foi processado com sucesso. O valor será
						creditado de acordo com o prazo do seu meio de
						pagamento.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Fatura:</strong> {invoiceId}
						</Text>
						<Text style={detailRow}>
							<strong>Valor reembolsado:</strong> {amount}
						</Text>
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
						<Text style={detailRow}>
							<strong>Motivo:</strong> {reason}
						</Text>
					</Section>

					<Text style={noteText}>
						Cartão de crédito: até 10 dias úteis. PIX: até 1 dia
						útil. Boleto: até 5 dias úteis.
					</Text>

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
	margin: "4px 0",
};
