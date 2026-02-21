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

interface InstanceConnectedEmailProps {
	userName?: string;
	instanceName?: string;
	phoneNumber?: string;
	connectedAt?: string;
	dashboardUrl?: string;
}

export default function InstanceConnectedEmail({
	userName = "Usuário",
	instanceName = "minha-instancia",
	phoneNumber = "+55 11 99999-0000",
	connectedAt,
	dashboardUrl = "https://zedaapi.com/instancias",
}: InstanceConnectedEmailProps) {
	const formattedDate = connectedAt
		? new Date(connectedAt).toLocaleDateString("pt-BR", {
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
			<Preview>WhatsApp conectado — {instanceName}</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>WhatsApp Conectado!</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Sua instância WhatsApp foi conectada com sucesso e está
						pronta para uso.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Instância:</strong> {instanceName}
						</Text>
						<Text style={detailRow}>
							<strong>Número:</strong> {phoneNumber}
						</Text>
						<Text style={detailRow}>
							<strong>Conectado em:</strong> {formattedDate}
						</Text>
						<Text style={detailRow}>
							<strong>Status:</strong> Online
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={dashboardUrl} style={ctaButton}>
							Gerenciar Instância
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
	backgroundColor: "#f0fdf4",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #bbf7d0",
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
