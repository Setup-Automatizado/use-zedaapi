import { NextResponse } from "next/server";
import { fetchCnpjData } from "@/lib/services/brasil-api/cnpj";
import {
	checkRateLimit,
	RATE_LIMIT_CONFIGS,
	validateOrigin,
	getClientIp,
	rateLimitHeaders,
} from "@/lib/rate-limit";

export async function GET(
	request: Request,
	{ params }: { params: Promise<{ cnpj: string }> },
) {
	if (!validateOrigin(request)) {
		return NextResponse.json({ error: "Acesso negado" }, { status: 403 });
	}

	const ip = getClientIp(request);
	const config = RATE_LIMIT_CONFIGS.API;
	const rl = await checkRateLimit(ip, config);

	if (!rl.success) {
		return NextResponse.json(
			{
				error: "Limite de requisições excedido. Tente novamente em alguns minutos.",
			},
			{ status: 429, headers: rateLimitHeaders(rl, config) },
		);
	}

	try {
		const { cnpj } = await params;

		// Format validation before proxying to external API
		const clean = cnpj.replace(/\D/g, "");
		if (!/^\d{14}$/.test(clean)) {
			return NextResponse.json(
				{ error: "CNPJ deve ter 14 dígitos numéricos" },
				{ status: 400 },
			);
		}

		const data = await fetchCnpjData(clean);
		return NextResponse.json(data, {
			headers: rateLimitHeaders(rl, config),
		});
	} catch (error) {
		const message =
			error instanceof Error ? error.message : "Erro ao consultar CNPJ";
		return NextResponse.json({ error: message }, { status: 400 });
	}
}
