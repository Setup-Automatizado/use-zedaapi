import type { Job } from "bullmq";
import type { EmailSendingJobData } from "@/lib/queue/types";
import { createLogger } from "@/lib/queue/logger";

const log = createLogger("processor:email");

export async function processEmailSendingJob(
	job: Job<EmailSendingJobData>,
): Promise<void> {
	const { to, template, data, attachments } = job.data;

	log.info("Processing email job", {
		jobId: job.id,
		to,
		template,
		hasAttachments: !!attachments?.length,
	});

	const done = log.timer("Email sending", { to, template });
	const { db } = await import("@/lib/db");

	// Load template from database
	const emailTemplate = await db.emailTemplate.findUnique({
		where: { slug: template },
	});

	if (!emailTemplate) {
		log.error("Email template not found", { template });
		throw new Error(`Email template "${template}" not found`);
	}

	if (!emailTemplate.active) {
		log.warn("Email template is inactive, skipping", { template });
		return;
	}

	// Render subject and body with template variables
	let subject = emailTemplate.subject;
	let body = emailTemplate.body;

	for (const [key, value] of Object.entries(data)) {
		const placeholder = `{{${key}}}`;
		const strValue = String(value ?? "");
		subject = subject.replaceAll(placeholder, strValue);
		body = body.replaceAll(placeholder, strValue);
	}

	// Send via nodemailer
	const nodemailer = await import("nodemailer");
	const transporter = nodemailer.createTransport({
		host: process.env.SMTP_HOST,
		port: Number(process.env.SMTP_PORT) || 587,
		secure: false,
		auth: {
			user: process.env.SMTP_USER,
			pass: process.env.SMTP_PASSWORD,
		},
	});

	const mailOptions: Record<string, unknown> = {
		from: process.env.SMTP_FROM,
		to,
		subject,
		html: body,
	};

	if (attachments?.length) {
		mailOptions.attachments = attachments.map((att) => ({
			filename: att.filename,
			content: att.content,
			encoding: att.encoding ?? "base64",
		}));
	}

	await transporter.sendMail(mailOptions);

	// Create notification record
	await db.notification.create({
		data: {
			userId: (data.userId as string) ?? "",
			type: template,
			channel: "email",
			subject,
			body,
			status: "sent",
			sentAt: new Date(),
			metadata: { to, template, jobId: job.id },
		},
	});

	done();
	log.info("Email sent successfully", { to, template });
}
