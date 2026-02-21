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

interface TwoFactorEnabledEmailProps {
	userName?: string;
	enabledAt?: string;
}

export default function TwoFactorEnabledEmail({
	userName = "Usuário",
	enabledAt,
}: TwoFactorEnabledEmailProps) {
	const formattedDate = enabledAt
		? new Date(enabledAt).toLocaleDateString("pt-BR", {
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
				Autenticação em dois fatores ativada — Zé da API Manager
			</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>2FA Ativado com Sucesso</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						A autenticação em dois fatores (2FA) foi ativada com
						sucesso na sua conta. A partir de agora, você precisará
						informar um código de verificação ao fazer login.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Data de ativação:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={tipBox}>
						<Text style={tipTitle}>Dica de segurança</Text>
						<Text style={tipText}>
							Guarde seus códigos de recuperação em um local
							seguro. Eles são a única forma de acessar sua conta
							caso perca o dispositivo autenticador.
						</Text>
					</Section>

					<Text style={warningText}>
						Se você não ativou o 2FA, entre em contato com nosso
						suporte imediatamente.
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
	color: "#16a34a",
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

const tipBox = {
	backgroundColor: "#f0fdf4",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "16px 0",
	border: "1px solid #bbf7d0",
};

const tipTitle = {
	color: "#15803d",
	fontSize: "14px",
	fontWeight: "600" as const,
	margin: "0 0 8px",
};

const tipText = {
	color: "#15803d",
	fontSize: "14px",
	lineHeight: "22px",
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
