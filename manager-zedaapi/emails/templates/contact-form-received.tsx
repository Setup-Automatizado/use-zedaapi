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

interface ContactFormReceivedEmailProps {
	userName?: string;
	ticketId?: string;
	subject?: string;
}

export default function ContactFormReceivedEmail({
	userName = "Usuário",
	ticketId = "SUP-2026-0042",
	subject = "Dúvida sobre integração",
}: ContactFormReceivedEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>Recebemos sua mensagem — Zé da API Manager</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>Mensagem Recebida</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Recebemos sua mensagem e nossa equipe irá analisá-la em
						breve. O prazo médio de resposta é de até 24 horas
						úteis.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Protocolo:</strong> {ticketId}
						</Text>
						<Text style={detailRow}>
							<strong>Assunto:</strong> {subject}
						</Text>
					</Section>

					<Text style={paragraph}>
						Responderemos diretamente neste e-mail. Por favor, não
						abra um novo chamado para o mesmo assunto.
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
