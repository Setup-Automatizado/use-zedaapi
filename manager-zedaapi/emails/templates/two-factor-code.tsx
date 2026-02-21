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

interface TwoFactorCodeEmailProps {
	userName?: string;
	code?: string;
	expiresIn?: string;
}

export default function TwoFactorCodeEmail({
	userName = "Usuário",
	code = "123456",
	expiresIn = "10 minutos",
}: TwoFactorCodeEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Seu código de verificação: {code}</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Código de Verificação</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Use o código abaixo para completar sua autenticação.
					</Text>

					<Section style={codeBox}>
						<Text style={codeText}>{code}</Text>
					</Section>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							Este código expira em <strong>{expiresIn}</strong>.
						</Text>
						<Text style={detailRow}>
							Se você não tentou fazer login, altere sua senha
							imediatamente.
						</Text>
					</Section>

					<Text style={warningText}>
						Nunca compartilhe este código com ninguém. Nossa equipe
						nunca pedirá seu código de verificação.
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

const codeBox = {
	backgroundColor: "#0f172a",
	borderRadius: "8px",
	padding: "20px",
	margin: "24px 0",
	textAlign: "center" as const,
};

const codeText = {
	color: "#ffffff",
	fontSize: "36px",
	fontWeight: "700" as const,
	letterSpacing: "8px",
	margin: "0",
	fontFamily: "monospace",
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
