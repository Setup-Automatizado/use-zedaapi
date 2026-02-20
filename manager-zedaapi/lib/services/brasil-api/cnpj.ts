"use server";

export interface CnpjData {
	razaoSocial: string;
	nomeFantasia: string | null;
	cnae: string | null;
	ie: string | null;
	// Address
	cep: string | null;
	logradouro: string | null;
	numero: string | null;
	complemento: string | null;
	bairro: string | null;
	municipio: string | null;
	uf: string | null;
	codigoMunicipio: string | null;
	// Contact
	email: string | null;
	telefone: string | null;
}

export async function fetchCnpjData(cnpj: string): Promise<CnpjData> {
	// Strip non-digits
	const cleanCnpj = cnpj.replace(/\D/g, "");

	if (cleanCnpj.length !== 14) {
		throw new Error("CNPJ deve ter 14 dígitos");
	}

	const response = await fetch(
		`https://brasilapi.com.br/api/cnpj/v1/${cleanCnpj}`,
		{
			next: { revalidate: 86400 }, // Cache 24h
		},
	);

	if (!response.ok) {
		if (response.status === 404) {
			throw new Error("CNPJ não encontrado");
		}
		throw new Error("Erro ao consultar CNPJ");
	}

	const data = await response.json();

	return {
		razaoSocial: data.razao_social || "",
		nomeFantasia: data.nome_fantasia || null,
		cnae: data.cnae_fiscal?.toString() || null,
		ie: null, // BrasilAPI doesn't provide IE
		cep: data.cep?.replace(/\D/g, "") || null,
		logradouro: data.logradouro || null,
		numero: data.numero || null,
		complemento: data.complemento || null,
		bairro: data.bairro || null,
		municipio: data.municipio || null,
		uf: data.uf || null,
		codigoMunicipio: data.codigo_municipio?.toString() || null,
		email: data.email || null,
		telefone: data.ddd_telefone_1 || null,
	};
}
