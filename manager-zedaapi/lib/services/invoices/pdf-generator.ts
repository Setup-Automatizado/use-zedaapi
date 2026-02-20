import { PDFDocument, StandardFonts, rgb, degrees } from "pdf-lib";

// =============================================================================
// INVOICE PDF GENERATOR - Manager ZedaAPI
// Pure JS (pdf-lib), Bun-compatible, no native deps.
// =============================================================================

interface InvoicePdfData {
	invoiceNumber: string;
	status: string;
	amount: number;
	currency: string;
	paymentMethod: string;
	paidAt: Date | null;
	createdAt: Date;
	dueDate?: Date | null;
	planName: string;
	user: { name: string; email: string };
	stripeInvoiceId?: string | null;
	sicrediChargeId?: string | null;
}

const C = {
	primary: rgb(0.16, 0.55, 0.94),
	brandBlue: rgb(0.22, 0.6, 0.97),
	headerBg: rgb(0.08, 0.09, 0.11),
	text: rgb(0.15, 0.15, 0.17),
	muted: rgb(0.5, 0.5, 0.55),
	light: rgb(0.7, 0.7, 0.73),
	sectionBg: rgb(0.975, 0.98, 0.98),
	white: rgb(1, 1, 1),
	border: rgb(0.88, 0.89, 0.9),
	success: rgb(0.1, 0.6, 0.42),
	warning: rgb(0.7, 0.5, 0.1),
	watermark: rgb(0.94, 0.95, 0.95),
};

const STATUS: Record<string, string> = {
	paid: "Pago",
	draft: "Processando",
	open: "Pendente",
	void: "Cancelada",
	uncollectible: "Inadimplente",
};

const METHOD: Record<string, string> = {
	pix: "PIX",
	card: "Cartao de Credito",
	boleto: "Boleto",
	stripe: "Stripe",
};

function fmtCurrency(value: number): string {
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(value);
}

function fmtDate(date: Date | string): string {
	return new Intl.DateTimeFormat("pt-BR", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
		hour: "2-digit",
		minute: "2-digit",
	}).format(new Date(date));
}

function fmtDateShort(date: Date | string): string {
	return new Intl.DateTimeFormat("pt-BR", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
	}).format(new Date(date));
}

function truncate(
	text: string,
	font: { widthOfTextAtSize: (t: string, s: number) => number },
	size: number,
	maxWidth: number,
): string {
	if (font.widthOfTextAtSize(text, size) <= maxWidth) return text;
	let t = text;
	while (t.length > 3 && font.widthOfTextAtSize(t + "...", size) > maxWidth) {
		t = t.slice(0, -1);
	}
	return t + "...";
}

// =============================================================================
// MAIN GENERATOR
// =============================================================================

export async function generateInvoicePdf(
	data: InvoicePdfData,
): Promise<Uint8Array> {
	const doc = await PDFDocument.create();
	const page = doc.addPage([595, 842]);
	const { width, height } = page.getSize();

	const regular = await doc.embedFont(StandardFonts.Helvetica);
	const bold = await doc.embedFont(StandardFonts.HelveticaBold);

	const LM = 48;
	const RM = width - 48;
	const CW = RM - LM;

	// Watermark
	const wmText = "ZEDAAPI";
	const wmSize = 72;
	const wmWidth = bold.widthOfTextAtSize(wmText, wmSize);
	page.drawText(wmText, {
		x: (width - wmWidth * 0.7) / 2,
		y: height / 2 - 30,
		size: wmSize,
		font: bold,
		color: C.watermark,
		rotate: degrees(-35),
		opacity: 0.5,
	});

	// Header
	const headerH = 110;
	page.drawRectangle({
		x: 0,
		y: height - headerH,
		width,
		height: headerH,
		color: C.headerBg,
	});
	page.drawRectangle({
		x: 0,
		y: height - headerH,
		width,
		height: 3,
		color: C.primary,
	});

	// Brand
	const brandX = LM;
	const brandY = height - 42;
	const zedaW = bold.widthOfTextAtSize("Zeda", 18);
	page.drawText("Zeda", {
		x: brandX,
		y: brandY,
		size: 18,
		font: bold,
		color: C.brandBlue,
	});
	page.drawText("API", {
		x: brandX + zedaW,
		y: brandY,
		size: 18,
		font: bold,
		color: C.white,
	});

	page.drawText("Comprovante de Pagamento", {
		x: brandX,
		y: height - 60,
		size: 10,
		font: regular,
		color: C.light,
	});

	// Invoice number (right)
	const invW = bold.widthOfTextAtSize(data.invoiceNumber, 13);
	page.drawText(data.invoiceNumber, {
		x: RM - invW,
		y: height - 36,
		size: 13,
		font: bold,
		color: C.white,
	});

	// Date (right)
	const dateStr = fmtDateShort(data.paidAt || data.createdAt);
	const dateW = regular.widthOfTextAtSize(dateStr, 9);
	page.drawText(dateStr, {
		x: RM - dateW,
		y: height - 52,
		size: 9,
		font: regular,
		color: C.light,
	});

	// Status badge (right)
	const statusText = STATUS[data.status] || data.status;
	const statusW = bold.widthOfTextAtSize(statusText, 8);
	const badgePad = 10;
	const badgeW = statusW + badgePad * 2;
	const badgeX = RM - badgeW;
	const badgeY = height - 76;
	page.drawRectangle({
		x: badgeX,
		y: badgeY,
		width: badgeW,
		height: 18,
		color: data.status === "paid" ? C.success : C.warning,
		borderWidth: 0,
	});
	page.drawText(statusText, {
		x: badgeX + badgePad,
		y: badgeY + 5,
		size: 8,
		font: bold,
		color: C.white,
	});

	// Content
	let y = height - headerH - 30;
	const labelX = LM + 12;
	const valueMaxW = CW * 0.55;

	function section(title: string) {
		y -= 8;
		page.drawRectangle({
			x: LM,
			y: y - 2,
			width: 3,
			height: 14,
			color: C.primary,
		});
		page.drawText(title, {
			x: LM + 10,
			y,
			size: 8,
			font: bold,
			color: C.primary,
		});
		y -= 8;
		page.drawLine({
			start: { x: LM, y },
			end: { x: RM, y },
			thickness: 0.5,
			color: C.border,
		});
		y -= 16;
	}

	function row(
		label: string,
		value: string,
		opts?: {
			valueBold?: boolean;
			valueColor?: ReturnType<typeof rgb>;
			valueSize?: number;
		},
	) {
		const vFont = opts?.valueBold ? bold : regular;
		const vSize = opts?.valueSize || 9;
		const vColor = opts?.valueColor || C.text;

		page.drawText(label, {
			x: labelX,
			y,
			size: 9,
			font: regular,
			color: C.muted,
		});

		const display = truncate(value, vFont, vSize, valueMaxW);
		const vW = vFont.widthOfTextAtSize(display, vSize);
		page.drawText(display, {
			x: RM - 12 - vW,
			y,
			size: vSize,
			font: vFont,
			color: vColor,
		});

		y -= 18;
	}

	// Plan
	section("PLANO");
	row("Nome", data.planName, { valueBold: true });

	// Client
	section("CLIENTE");
	row("Nome", data.user.name, { valueBold: true });
	row("Email", data.user.email);

	// Value
	section("VALOR");
	row("Valor", fmtCurrency(data.amount), {
		valueBold: true,
		valueColor: C.success,
		valueSize: 12,
	});

	// Payment
	section("PAGAMENTO");
	row("Metodo", METHOD[data.paymentMethod] || data.paymentMethod);
	if (data.stripeInvoiceId) {
		row("Stripe ID", data.stripeInvoiceId);
	}
	if (data.sicrediChargeId) {
		row("Sicredi ID", data.sicrediChargeId);
	}
	if (data.paidAt) {
		row("Pago em", fmtDate(data.paidAt));
	}
	if (data.dueDate) {
		row("Vencimento", fmtDateShort(data.dueDate));
	}

	// Footer
	const footerTop = 65;
	page.drawLine({
		start: { x: LM, y: footerTop },
		end: { x: RM, y: footerTop },
		thickness: 0.5,
		color: C.border,
	});

	const fBrandY = footerTop - 18;
	const fZedaW = bold.widthOfTextAtSize("Zeda", 8);
	page.drawText("Zeda", {
		x: LM,
		y: fBrandY,
		size: 8,
		font: bold,
		color: C.brandBlue,
	});
	page.drawText("API", {
		x: LM + fZedaW,
		y: fBrandY,
		size: 8,
		font: bold,
		color: C.muted,
	});

	page.drawText(
		"Comprovante gerado automaticamente pela plataforma Zé da API Manager.",
		{
			x: LM,
			y: footerTop - 34,
			size: 7,
			font: regular,
			color: C.light,
		},
	);

	const genStr = `Gerado: ${fmtDate(new Date())}`;
	const genW = regular.widthOfTextAtSize(genStr, 7);
	page.drawText(genStr, {
		x: RM - genW,
		y: footerTop - 18,
		size: 7,
		font: regular,
		color: C.light,
	});

	const idW = regular.widthOfTextAtSize(data.invoiceNumber, 7);
	page.drawText(data.invoiceNumber, {
		x: RM - idW,
		y: footerTop - 34,
		size: 7,
		font: regular,
		color: C.light,
	});

	return doc.save();
}
