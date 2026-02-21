"use server";

import { z } from "zod";
import { headers } from "next/headers";
import { sendEmail } from "@/lib/email";
import { sendWebhook } from "@/lib/webhook";
import type { ActionResult } from "@/types";

// =============================================================================
// Validation schemas
// =============================================================================

const contactSchema = z.object({
	nome: z.string().min(2, "Nome deve ter pelo menos 2 caracteres"),
	email: z.string().email("E-mail invalido"),
	whatsapp: z
		.string()
		.regex(/^\(?\d{2}\)?\s?\d{4,5}-?\d{4}$/, "Formato de WhatsApp invalido")
		.or(z.literal(""))
		.optional(),
	empresa: z.string().optional(),
	assunto: z.enum([
		"comercial",
		"suporte",
		"parceria",
		"financeiro",
		"outro",
	]),
	mensagem: z.string().min(10, "Mensagem deve ter pelo menos 10 caracteres"),
	preferencia_contato: z.enum(["email", "whatsapp", "ambos"]),
	utm_source: z.string().optional(),
	utm_medium: z.string().optional(),
	utm_campaign: z.string().optional(),
	utm_term: z.string().optional(),
	utm_content: z.string().optional(),
	page_url: z.string().optional(),
	referrer: z.string().optional(),
});

const widgetSchema = z.object({
	nome: z.string().min(2, "Nome deve ter pelo menos 2 caracteres"),
	preferencia: z.enum(["email", "whatsapp"]),
	contato: z.string().min(3, "Informe um e-mail ou WhatsApp valido"),
	mensagem: z
		.string()
		.min(3, "Mensagem deve ter pelo menos 3 caracteres")
		.optional()
		.or(z.literal("")),
	page_url: z.string().optional(),
	referrer: z.string().optional(),
});

// =============================================================================
// Constants
// =============================================================================

const ASSUNTO_LABELS: Record<string, string> = {
	comercial: "Comercial",
	suporte: "Suporte Tecnico",
	parceria: "Parceria",
	financeiro: "Financeiro",
	outro: "Outro",
};

// =============================================================================
// Email template
// =============================================================================

function buildConfirmationEmail(nome: string, assuntoLabel: string): string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8" /><meta name="viewport" content="width=device-width,initial-scale=1.0" /></head>
<body style="margin:0;padding:0;background-color:#fafafa;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background-color:#fafafa;padding:40px 16px;">
<tr><td align="center">
<table width="560" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.08);">
  <tr><td style="background-color:#16a34a;padding:24px 32px;">
    <h1 style="margin:0;color:#ffffff;font-size:18px;font-weight:600;letter-spacing:-0.02em;">Zé da API</h1>
  </td></tr>
  <tr><td style="padding:32px;">
    <h2 style="margin:0 0 16px;color:#18181b;font-size:20px;font-weight:600;">Ola, ${nome}!</h2>
    <p style="margin:0 0 16px;color:#3f3f46;font-size:14px;line-height:1.7;">
      Recebemos sua mensagem sobre <strong style="color:#18181b;">${assuntoLabel}</strong> e nossa equipe ja esta analisando. Voce recebera um retorno em ate <strong style="color:#18181b;">24 horas uteis</strong>.
    </p>
    <p style="margin:0 0 24px;color:#3f3f46;font-size:14px;line-height:1.7;">
      Se precisar de ajuda imediata, fale diretamente pelo WhatsApp:
    </p>
    <table cellpadding="0" cellspacing="0"><tr><td>
      <a href="https://wa.me/5521971532700" style="display:inline-block;background-color:#16a34a;color:#ffffff;font-size:14px;font-weight:600;text-decoration:none;padding:12px 24px;border-radius:9999px;">Falar no WhatsApp</a>
    </td></tr></table>
  </td></tr>
  <tr><td style="padding:0 32px 32px;">
    <hr style="border:none;border-top:1px solid #e4e4e7;margin:0 0 16px;" />
    <p style="margin:0;color:#a1a1aa;font-size:12px;line-height:1.5;">
      Equipe Zé da API &mdash; contato@zedaapi.com
    </p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>`;
}

// =============================================================================
// Server Actions
// =============================================================================

export async function submitContactForm(
	formData: FormData,
): Promise<ActionResult> {
	const raw = {
		nome: formData.get("nome"),
		email: formData.get("email"),
		whatsapp: formData.get("whatsapp"),
		empresa: formData.get("empresa"),
		assunto: formData.get("assunto"),
		mensagem: formData.get("mensagem"),
		preferencia_contato: formData.get("preferencia_contato"),
		utm_source: formData.get("utm_source"),
		utm_medium: formData.get("utm_medium"),
		utm_campaign: formData.get("utm_campaign"),
		utm_term: formData.get("utm_term"),
		utm_content: formData.get("utm_content"),
		page_url: formData.get("page_url"),
		referrer: formData.get("referrer"),
	};

	const validation = contactSchema.safeParse(raw);
	if (!validation.success) {
		return {
			success: false,
			errors: validation.error.flatten().fieldErrors,
		};
	}

	const data = validation.data;
	const headersList = await headers();

	await sendWebhook({
		event: "contact_form",
		timestamp: new Date().toISOString(),
		data: {
			nome: data.nome,
			email: data.email,
			whatsapp: data.whatsapp || null,
			empresa: data.empresa || null,
			assunto: data.assunto,
			assunto_label: ASSUNTO_LABELS[data.assunto] ?? data.assunto,
			mensagem: data.mensagem,
			preferencia_contato: data.preferencia_contato,
		},
		utm: {
			source: data.utm_source || null,
			medium: data.utm_medium || null,
			campaign: data.utm_campaign || null,
			term: data.utm_term || null,
			content: data.utm_content || null,
		},
		metadata: {
			page_url: data.page_url || null,
			referrer: data.referrer || null,
			user_agent: headersList.get("user-agent"),
		},
	});

	try {
		const assuntoLabel = ASSUNTO_LABELS[data.assunto] ?? data.assunto;
		await sendEmail(
			data.email,
			"Recebemos seu contato - Zé da API",
			buildConfirmationEmail(data.nome, assuntoLabel),
		);
	} catch {
		// Email failure should not block form submission
	}

	return { success: true };
}

export async function submitWidgetForm(
	formData: FormData,
): Promise<ActionResult> {
	const raw = {
		nome: formData.get("nome"),
		preferencia: formData.get("preferencia"),
		contato: formData.get("contato"),
		mensagem: formData.get("mensagem"),
		page_url: formData.get("page_url"),
		referrer: formData.get("referrer"),
	};

	const validation = widgetSchema.safeParse(raw);
	if (!validation.success) {
		return {
			success: false,
			errors: validation.error.flatten().fieldErrors,
		};
	}

	const data = validation.data;
	const headersList = await headers();

	await sendWebhook({
		event: "whatsapp_widget",
		timestamp: new Date().toISOString(),
		data: {
			nome: data.nome,
			preferencia: data.preferencia,
			contato: data.contato,
			mensagem: data.mensagem || null,
		},
		utm: {
			source: null,
			medium: null,
			campaign: null,
			term: null,
			content: null,
		},
		metadata: {
			page_url: data.page_url || null,
			referrer: data.referrer || null,
			user_agent: headersList.get("user-agent"),
		},
	});

	return { success: true };
}
