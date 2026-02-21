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

interface LoginNewDeviceEmailProps {
	userName?: string;
	device?: string;
	browser?: string;
	location?: string;
	ipAddress?: string;
	loginAt?: string;
	supportUrl?: string;
}

export default function LoginNewDeviceEmail({
	userName = "Usuário",
	device = "MacBook Pro",
	browser = "Chrome 120",
	location = "São Paulo, SP, Brasil",
	ipAddress = "187.45.123.78",
	loginAt,
	supportUrl = "https://zedaapi.com/suporte",
}: LoginNewDeviceEmailProps) {
	const formattedDate = loginAt
		? new Date(loginAt).toLocaleDateString("pt-BR", {
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
			<Preview>Novo login detectado — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Novo Login Detectado</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Detectamos um novo login na sua conta a partir de um
						dispositivo que não reconhecemos.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Dispositivo:</strong> {device}
						</Text>
						<Text style={detailRow}>
							<strong>Navegador:</strong> {browser}
						</Text>
						<Text style={detailRow}>
							<strong>Localização:</strong> {location}
						</Text>
						<Text style={detailRow}>
							<strong>IP:</strong> {ipAddress}
						</Text>
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
					</Section>

					<Section style={warningBox}>
						<Text style={warningTitle}>
							Não reconhece este acesso?
						</Text>
						<Text style={warningBody}>
							Se você não realizou este login, altere sua senha
							imediatamente e entre em contato com nosso suporte.
						</Text>
					</Section>

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
	color: "#d97706",
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
	backgroundColor: "#fffbeb",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #fde68a",
};

const detailRow = {
	color: "#334155",
	fontSize: "14px",
	lineHeight: "28px",
	margin: "0",
};

const warningBox = {
	backgroundColor: "#fef2f2",
	borderRadius: "8px",
	padding: "16px 24px",
	margin: "16px 0",
	border: "1px solid #fecaca",
};

const warningTitle = {
	color: "#b91c1c",
	fontSize: "14px",
	fontWeight: "600" as const,
	margin: "0 0 8px",
};

const warningBody = {
	color: "#b91c1c",
	fontSize: "14px",
	lineHeight: "22px",
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
