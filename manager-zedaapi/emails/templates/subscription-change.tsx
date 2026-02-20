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

interface SubscriptionChangeEmailProps {
	userName?: string;
	oldPlan?: string;
	newPlan?: string;
	effectiveDate?: string;
	newPrice?: string;
	action?: "upgrade" | "downgrade" | "cancel" | "resume";
}

export default function SubscriptionChangeEmail({
	userName = "Usuário",
	oldPlan = "—",
	newPlan = "—",
	effectiveDate,
	newPrice = "—",
	action = "upgrade",
}: SubscriptionChangeEmailProps) {
	const titles: Record<string, string> = {
		upgrade: "Upgrade de Plano",
		downgrade: "Downgrade de Plano",
		cancel: "Cancelamento de Assinatura",
		resume: "Reativação de Assinatura",
	};

	const descriptions: Record<string, string> = {
		upgrade: `Seu plano foi atualizado de ${oldPlan} para ${newPlan}.`,
		downgrade: `Seu plano será alterado de ${oldPlan} para ${newPlan} ao final do período atual.`,
		cancel: "Sua assinatura será cancelada ao final do período atual. Suas instâncias continuarão funcionando até lá.",
		resume: `Sua assinatura foi reativada no plano ${newPlan}.`,
	};

	const formattedDate = effectiveDate
		? new Date(effectiveDate).toLocaleDateString("pt-BR")
		: "—";

	return (
		<Html>
			<Head />
			<Preview>Zé da API — {titles[action] ?? action}</Preview>
			<Body style={main}>
				<Container style={container}>
					<Heading style={logo}>Zé da API Manager</Heading>
					<Hr style={hr} />
					<Heading style={heading}>
						{titles[action] ?? action}
					</Heading>
					<Text style={paragraph}>Olá {userName},</Text>
					<Text style={paragraph}>{descriptions[action] ?? ""}</Text>

					<Section style={detailsBox}>
						{action !== "cancel" && (
							<>
								<Text style={detailRow}>
									<strong>Plano anterior:</strong> {oldPlan}
								</Text>
								<Text style={detailRow}>
									<strong>Novo plano:</strong> {newPlan}
								</Text>
								<Text style={detailRow}>
									<strong>Novo valor:</strong> {newPrice}
								</Text>
							</>
						)}
						<Text style={detailRow}>
							<strong>Data efetiva:</strong> {formattedDate}
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
