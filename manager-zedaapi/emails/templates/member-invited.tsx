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

interface MemberInvitedEmailProps {
	userName?: string;
	organizationName?: string;
	invitedBy?: string;
	role?: string;
	acceptUrl?: string;
}

export default function MemberInvitedEmail({
	userName = "Usuário",
	organizationName = "Empresa ABC",
	invitedBy = "Carlos Souza",
	role = "Membro",
	acceptUrl = "https://zedaapi.com/organizacao/convite?token=abc123",
}: MemberInvitedEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>
				Convite para {organizationName} — Zé da API Manager
			</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Convite para Organização</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						<strong>{invitedBy}</strong> convidou você para fazer
						parte da organização <strong>{organizationName}</strong>{" "}
						no Zé da API Manager.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Organização:</strong> {organizationName}
						</Text>
						<Text style={detailRow}>
							<strong>Convidado por:</strong> {invitedBy}
						</Text>
						<Text style={detailRow}>
							<strong>Função:</strong> {role}
						</Text>
					</Section>

					<Section style={ctaSection}>
						<Link href={acceptUrl} style={ctaButton}>
							Aceitar Convite
						</Link>
					</Section>

					<Text style={noteText}>
						Este convite expira em 7 dias. Caso não reconheça este
						convite, ignore este e-mail.
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
	margin: "20px 0 0",
};
