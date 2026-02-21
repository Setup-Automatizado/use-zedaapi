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

interface PaymentMethodUpdatedEmailProps {
	userName?: string;
	methodType?: string;
	lastFourDigits?: string;
	updatedAt?: string;
}

export default function PaymentMethodUpdatedEmail({
	userName = "Usuário",
	methodType = "Cartão de Crédito",
	lastFourDigits = "4242",
	updatedAt,
}: PaymentMethodUpdatedEmailProps) {
	const formattedDate = updatedAt
		? new Date(updatedAt).toLocaleDateString("pt-BR", {
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
			<Preview>
				Método de pagamento atualizado — Zé da API Manager
			</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>
						Método de Pagamento Atualizado
					</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Seu método de pagamento foi atualizado com sucesso.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Tipo:</strong> {methodType}
						</Text>
						{lastFourDigits && (
							<Text style={detailRow}>
								<strong>Final:</strong> ****{lastFourDigits}
							</Text>
						)}
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
					</Section>

					<Text style={warningText}>
						Se você não realizou essa alteração, entre em contato
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
