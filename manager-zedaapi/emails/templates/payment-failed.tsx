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

interface PaymentFailedEmailProps {
	userName?: string;
	amount?: string;
	dueDate?: string;
	retryUrl?: string;
}

export default function PaymentFailedEmail({
	userName = "Usuário",
	amount = "R$ 0,00",
	dueDate,
	retryUrl = "https://zedaapi.com/faturamento",
}: PaymentFailedEmailProps) {
	const formattedDate = dueDate
		? new Date(dueDate).toLocaleDateString("pt-BR")
		: "—";

	return (
		<Html>
			<Head />
			<Preview>Falha no pagamento — Zé da API</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Falha no Pagamento</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Não conseguimos processar o pagamento da sua assinatura.
						Por favor, atualize seu método de pagamento para evitar
						a suspensão do serviço.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Valor:</strong> {amount}
						</Text>
						<Text style={detailRow}>
							<strong>Vencimento:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={retryUrl} style={ctaButton}>
							Atualizar Pagamento
						</Link>
					</Section>

					<Text style={warningText}>
						Caso o pagamento não seja regularizado, suas instâncias
						poderão ser desconectadas.
					</Text>

					<Hr style={hr} />
					<Text style={footer}>
						Precisa de ajuda? Entre em contato com nosso suporte.
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
	margin: "4px 0",
};
