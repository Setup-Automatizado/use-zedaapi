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

interface MemberJoinedEmailProps {
	userName?: string;
	memberName?: string;
	memberEmail?: string;
	organizationName?: string;
	role?: string;
	joinedAt?: string;
}

export default function MemberJoinedEmail({
	userName = "Administrador",
	memberName = "Ana Costa",
	memberEmail = "ana@exemplo.com",
	organizationName = "Empresa ABC",
	role = "Membro",
	joinedAt,
}: MemberJoinedEmailProps) {
	const formattedDate = joinedAt
		? new Date(joinedAt).toLocaleDateString("pt-BR")
		: new Date().toLocaleDateString("pt-BR");

	return (
		<Html>
			<Head />
			<Preview>
				{memberName} entrou na organização — Zé da API Manager
			</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Novo Membro</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						<strong>{memberName}</strong> aceitou o convite e agora
						faz parte da organização{" "}
						<strong>{organizationName}</strong>.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Membro:</strong> {memberName}
						</Text>
						<Text style={detailRow}>
							<strong>E-mail:</strong> {memberEmail}
						</Text>
						<Text style={detailRow}>
							<strong>Função:</strong> {role}
						</Text>
						<Text style={detailRow}>
							<strong>Data:</strong> {formattedDate}
						</Text>
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
