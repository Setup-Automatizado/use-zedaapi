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

interface NfseIssuedEmailProps {
	userName?: string;
	nfseNumber?: string;
	amount?: string;
	pdfUrl?: string | null;
}

export default function NfseIssuedEmail({
	userName = "Usuário",
	nfseNumber = "",
	amount = "R$ 0,00",
	pdfUrl,
}: NfseIssuedEmailProps) {
	return (
		<Html>
			<Head />
			<Preview>NFS-e emitida — {nfseNumber}</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>NFS-e Emitida</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>
						Sua Nota Fiscal de Serviço Eletrônica (NFS-e) foi
						emitida com sucesso.
					</Text>

					<Section style={detailsBox}>
						<Text style={detailRow}>
							<strong>Número NFS-e:</strong> {nfseNumber}
						</Text>
						<Text style={detailRow}>
							<strong>Valor:</strong> {amount}
						</Text>
					</Section>

					{pdfUrl && (
						<Section style={ctaSection}>
							<Link href={pdfUrl} style={ctaButton}>
								Baixar DANFSE
							</Link>
						</Section>
					)}

					<Hr style={hr} />
					<Text style={footer}>
						Este documento fiscal tem validade jurídica conforme a
						legislação vigente.
					</Text>
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

const hr = {
	borderColor: "#e2e8f0",
	margin: "20px 0",
};

const footer = {
	color: "#94a3b8",
	fontSize: "12px",
	textAlign: "center" as const,
	margin: "4px 0",
};
