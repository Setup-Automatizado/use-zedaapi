import { NextResponse } from "next/server";
import { fetchCepData } from "@/lib/services/brasil-api/cep";
import {
	checkRateLimit,
	RATE_LIMIT_CONFIGS,
	validateOrigin,
	getClientIp,
	rateLimitHeaders,
} from "@/lib/rate-limit";

export async function GET(
	request: Request,
	{ params }: { params: Promise<{ cep: string }> },
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
		const { cep } = await params;

		// Format validation before proxying to external API
		const clean = cep.replace(/\D/g, "");
		if (!/^\d{8}$/.test(clean)) {
			return NextResponse.json(
				{ error: "CEP deve ter 8 dígitos numéricos" },
				{ status: 400 },
			);
		}

		const data = await fetchCepData(clean);
		return NextResponse.json(data, {
			headers: rateLimitHeaders(rl, config),
		});
	} catch (error) {
		const message =
			error instanceof Error ? error.message : "Erro ao consultar CEP";
		return NextResponse.json({ error: message }, { status: 400 });
	}
}
