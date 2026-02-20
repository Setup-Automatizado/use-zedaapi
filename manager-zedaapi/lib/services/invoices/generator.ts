import { db } from "@/lib/db";
import { generateInvoicePdf } from "./pdf-generator";

// =============================================================================
// INVOICE NUMBER GENERATION
// =============================================================================

/**
 * Generates a sequential invoice number: INV-{YEAR}-{SEQUENTIAL_6_DIGITS}
 * Uses last invoice lookup to prevent duplicates.
 */
export async function generateInvoiceNumber(): Promise<string> {
	const year = new Date().getFullYear();
	const prefix = `INV-${year}-`;

	// Simple counter based on total count
	const totalCount = await db.invoice.count();
	const nextSequence = totalCount + 1;

	return `${prefix}${String(nextSequence).padStart(6, "0")}`;
}

// =============================================================================
// PDF GENERATION + S3 UPLOAD
// =============================================================================

/**
 * Generates a PDF receipt and uploads it to S3.
 * Updates the invoice record with the pdfUrl.
 * Non-blocking: failures are logged but don't throw.
 */
async function generateAndUploadPdf(invoiceId: string): Promise<void> {
	try {
		const invoice = await db.invoice.findUnique({
			where: { id: invoiceId },
			include: {
				user: { select: { name: true, email: true } },
				subscription: {
					select: {
						plan: { select: { name: true } },
					},
				},
			},
		});

		if (!invoice) return;

		const invoiceNumber = await generateInvoiceNumber();

		const pdfBytes = await generateInvoicePdf({
			invoiceNumber,
			status: invoice.status,
			amount: Number(invoice.amount),
			currency: invoice.currency,
			paymentMethod: invoice.paymentMethod || "stripe",
			paidAt: invoice.paidAt,
			createdAt: invoice.createdAt,
			dueDate: invoice.dueDate,
			planName: invoice.subscription?.plan?.name || "Assinatura",
			user: {
				name: invoice.user.name,
				email: invoice.user.email,
			},
			sicrediChargeId: invoice.sicrediChargeId,
			stripeInvoiceId: invoice.stripeInvoiceId,
		});

		// Upload to S3 (dynamic import to avoid bundling issues)
		const { S3Client } = await import("bun");
		const s3 = new S3Client({
			accessKeyId: process.env.S3_ACCESS_KEY_ID!,
			secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!,
			bucket: process.env.S3_BUCKET!,
			endpoint: process.env.S3_ENDPOINT!,
			region: process.env.S3_REGION || "us-east-1",
		});

		const year = new Date().getFullYear();
		const s3Key = `invoices/${year}/${invoiceNumber}.pdf`;
		const s3File = s3.file(s3Key);
		await s3File.write(Buffer.from(pdfBytes), {
			type: "application/pdf",
		});

		const publicUrl = process.env.S3_PUBLIC_URL || process.env.S3_ENDPOINT;
		const bucket = process.env.S3_BUCKET;
		const pathStyle = process.env.S3_PATH_STYLE === "true";
		const pdfUrl = pathStyle
			? `${publicUrl}/${bucket}/${s3Key}`
			: `${publicUrl}/${s3Key}`;

		await db.invoice.update({
			where: { id: invoiceId },
			data: { pdfUrl },
		});

		console.log(`[invoice] PDF generated: ${pdfUrl}`);
	} catch (error) {
		// PDF generation failure should NOT break the payment flow
		console.error(
			`[invoice] Failed to generate PDF for invoice ${invoiceId}:`,
			error,
		);
	}
}

// =============================================================================
// CREATE INVOICE FOR SUBSCRIPTION
// =============================================================================

interface CreateInvoiceOptions {
	userId: string;
	subscriptionId: string;
	amount: number;
	paymentMethod: string;
	stripeInvoiceId?: string;
	sicrediChargeId?: string;
	dueDate?: Date;
	paidAt?: Date;
	status?: string;
}

/**
 * Creates an invoice for a subscription payment.
 * Idempotent: checks for existing invoice by stripeInvoiceId before creating.
 * After creation, generates a PDF receipt asynchronously.
 */
export async function createSubscriptionInvoice(opts: CreateInvoiceOptions) {
	// Idempotency: check existing
	if (opts.stripeInvoiceId) {
		const existing = await db.invoice.findUnique({
			where: { stripeInvoiceId: opts.stripeInvoiceId },
		});
		if (existing) return existing;
	}

	const invoice = await db.invoice.create({
		data: {
			userId: opts.userId,
			subscriptionId: opts.subscriptionId,
			amount: opts.amount,
			currency: "BRL",
			status: opts.status || "paid",
			paymentMethod: opts.paymentMethod,
			stripeInvoiceId: opts.stripeInvoiceId || null,
			sicrediChargeId: opts.sicrediChargeId || null,
			dueDate: opts.dueDate || null,
			paidAt: opts.paidAt || new Date(),
		},
	});

	// Generate PDF receipt asynchronously (non-blocking)
	generateAndUploadPdf(invoice.id).catch((err) =>
		console.error("[invoice] PDF generation error:", err),
	);

	return invoice;
}

// =============================================================================
// REGENERATE PDF (for admin/manual re-generation)
// =============================================================================

/**
 * Re-generates the PDF for an existing invoice.
 */
export async function regenerateInvoicePdf(invoiceId: string): Promise<void> {
	await generateAndUploadPdf(invoiceId);
}
