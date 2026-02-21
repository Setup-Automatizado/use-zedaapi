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

interface InstanceCreatedEmailProps {
	userName?: string;
	instanceName?: string;
	instanceId?: string;
	dashboardUrl?: string;
}

export default function InstanceCreatedEmail({
	userName = "Usuário",
	instanceName = "minha-instancia",
	instanceId = "inst_abc123def456",
	dashboardUrl = "https://zedaapi.com/instancias",
}: InstanceCreatedEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Instância criada — {instanceName}</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Instância Criada!</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Sua nova instância WhatsApp foi provisionada com
						sucesso. Agora você precisa conectá-la escaneando o QR
						Code.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Nome:</strong> {instanceName}
						</Text>
						<Text style={detailRow}>
							<strong>ID:</strong> {instanceId}
						</Text>
						<Text style={detailRow}>
							<strong>Status:</strong> Aguardando conexão
						</Text>
					</Section>

					<Section style={steps}>
						<Text style={stepTitle}>Próximos passos:</Text>
						<Text style={step}>
							1. Acesse o painel e abra a instância
						</Text>
						<Text style={step}>
							2. Escaneie o QR Code com o WhatsApp
						</Text>
						<Text style={step}>3. Configure seus webhooks</Text>
						<Text style={step}>
							4. Comece a enviar mensagens via API
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={dashboardUrl} style={ctaButton}>
							Conectar Instância
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

const steps = {
	backgroundColor: "#f0fdf4",
	borderRadius: "8px",
	padding: "20px 24px",
	margin: "20px 0",
	border: "1px solid #bbf7d0",
};

const stepTitle = {
	color: "#0f172a",
	fontSize: "14px",
	fontWeight: "600" as const,
	margin: "0 0 12px",
};

const step = {
	color: "#475569",
	fontSize: "14px",
	lineHeight: "24px",
	margin: "4px 0",
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
