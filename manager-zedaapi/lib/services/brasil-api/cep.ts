"use server";

export interface CepData {
	logradouro: string | null;
	bairro: string | null;
	localidade: string;
	uf: string;
	complemento: string | null;
	ibge: string | null;
}

export async function fetchCepData(cep: string): Promise<CepData> {
	const cleanCep = cep.replace(/\D/g, "");

	if (cleanCep.length !== 8) {
		throw new Error("CEP deve ter 8 dígitos");
	}

	const response = await fetch(`https://viacep.com.br/ws/${cleanCep}/json/`, {
		next: { revalidate: 604800 }, // Cache 7 days
	});

	if (!response.ok) {
		throw new Error("Erro ao consultar CEP");
	}

	const data = await response.json();

	if (data.erro) {
		throw new Error("CEP não encontrado");
	}

	return {
		logradouro: data.logradouro || null,
		bairro: data.bairro || null,
		localidade: data.localidade || "",
		uf: data.uf || "",
		complemento: data.complemento || null,
		ibge: data.ibge || null,
	};
}
