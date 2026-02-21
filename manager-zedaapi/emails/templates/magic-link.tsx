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

interface MagicLinkEmailProps {
	userName?: string;
	magicLinkUrl?: string;
	expiresIn?: string;
}

export default function MagicLinkEmail({
	userName = "Usuário",
	magicLinkUrl = "https://zedaapi.com/auth/magic-link?token=abc123",
	expiresIn = "15 minutos",
}: MagicLinkEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Seu link de acesso — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Link de Acesso</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Você solicitou um link de acesso para entrar na sua
						conta. Clique no botão abaixo para fazer login.
					</Text>

					<Section style={ctaSection}>
						<Link href={magicLinkUrl} style={ctaButton}>
							Entrar na Minha Conta
						</Link>
					</Section>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							Este link expira em <strong>{expiresIn}</strong>.
						</Text>
						<Text style={detailRow}>
							Se você não solicitou este link, ignore este e-mail.
						</Text>
					</Section>

					<Text style={warningText}>
						Nunca compartilhe este link. Nossa equipe nunca pedirá
						seu link de acesso.
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
