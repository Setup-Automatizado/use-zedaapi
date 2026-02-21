import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { sendEmail } from "@/lib/email";
import { sendWebhook } from "@/lib/webhook";

export const dynamic = "force-dynamic";

const deletionSchema = z.object({
	nome: z.string().min(2, "Nome deve ter pelo menos 2 caracteres"),
	email: z.string().email("E-mail inválido"),
	cpf_cnpj: z.string().min(11).max(18).optional(),
	motivo: z.string().min(5).optional(),
});

function buildDeletionEmail(nome: string): string {
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
      Recebemos sua solicitacao de exclusao de dados. Processaremos em ate <strong style="color:#18181b;">15 dias uteis</strong> conforme a LGPD (Lei Geral de Protecao de Dados).
    </p>
    <p style="margin:0 0 24px;color:#3f3f46;font-size:14px;line-height:1.7;">
      Caso tenha duvidas, entre em contato pelo e-mail
      <a href="mailto:privacidade@zedaapi.com" style="color:#16a34a;font-weight:500;">privacidade@zedaapi.com</a>.
    </p>
  </td></tr>
  <tr><td style="padding:0 32px 32px;">
    <hr style="border:none;border-top:1px solid #e4e4e7;margin:0 0 16px;" />
    <p style="margin:0;color:#a1a1aa;font-size:12px;line-height:1.5;">
      Equipe Zé da API &mdash; Protecao de Dados
    </p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>`;
}

export async function POST(request: NextRequest) {
	let body: unknown;
	try {
		body = await request.json();
	} catch {
		return NextResponse.json(
			{ error: "Invalid JSON body" },
			{ status: 400 },
		);
	}

	const validation = deletionSchema.safeParse(body);
	if (!validation.success) {
		return NextResponse.json(
			{
				error: "Validation failed",
				details: validation.error.flatten().fieldErrors,
			},
			{ status: 422 },
		);
	}

	const data = validation.data;

	await sendWebhook({
		event: "data_deletion_request",
		timestamp: new Date().toISOString(),
		data: {
			nome: data.nome,
			email: data.email,
			cpf_cnpj: data.cpf_cnpj ?? null,
			motivo: data.motivo ?? null,
		},
		utm: {
			source: null,
			medium: null,
			campaign: null,
			term: null,
			content: null,
		},
		metadata: {
			page_url: request.headers.get("referer"),
			referrer: null,
			user_agent: request.headers.get("user-agent"),
		},
	});

	try {
		await sendEmail(
			data.email,
			"Solicitacao de exclusao de dados recebida - Zé da API",
			buildDeletionEmail(data.nome),
		);
	} catch {
		// Email failure should not block the response
	}

	return NextResponse.json({
		success: true,
		message:
			"Solicitacao de exclusao de dados recebida. Processaremos em ate 15 dias uteis conforme a LGPD.",
	});
}
