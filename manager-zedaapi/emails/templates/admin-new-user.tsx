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

interface AdminNewUserEmailProps {
	adminName?: string;
	userName?: string;
	userEmail?: string;
	registeredAt?: string;
	adminUrl?: string;
}

export default function AdminNewUserEmail({
	adminName = "Administrador",
	userName = "João Silva",
	userEmail = "joao@exemplo.com",
	registeredAt,
	adminUrl = "https://zedaapi.com/admin/usuarios",
}: AdminNewUserEmailProps) {
	const formattedDate = registeredAt
		? new Date(registeredAt).toLocaleDateString("pt-BR", {
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
			<Preview>Novo usuário: {userName} — Admin Zé da API</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Novo Usuário Registrado</Heading>
					<Text style={paragraph}>Olá {adminName},</Text>
					<Text style={paragraph}>
						Um novo usuário se registrou na plataforma.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Nome:</strong> {userName}
						</Text>
						<Text style={detailRow}>
							<strong>E-mail:</strong> {userEmail}
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
