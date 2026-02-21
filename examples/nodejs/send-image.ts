/**
 * Zé da API — Enviar imagem
 * Documentação: https://api.zedaapi.com/docs
 *
 * Uso: npx tsx send-image.ts
 */

// Configuração — substitua pelos seus dados
const HOST = process.env.ZEDAAPI_HOST || "https://sua-instancia.zedaapi.com";
const INSTANCE_ID = process.env.ZEDAAPI_INSTANCE_ID || "sua-instancia";
const INSTANCE_TOKEN = process.env.ZEDAAPI_INSTANCE_TOKEN || "seu-token-aqui";

// URL base com autenticação embutida na rota
const BASE_URL = `${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}`;

interface SendImageResponse {
  key?: {
    remoteJid: string;
    fromMe: boolean;
    id: string;
  };
  message?: Record<string, unknown>;
  error?: string;
}

async function sendImage(
  phone: string,
  imageUrl: string,
  caption?: string,
): Promise<SendImageResponse> {
  const response = await fetch(`${BASE_URL}/send-image`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ phone, image: imageUrl, caption }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`Erro ${response.status}: ${error}`);
  }

  return response.json() as Promise<SendImageResponse>;
}

// Exemplo de uso
async function main() {
  try {
    const result = await sendImage(
      "5511999999999",
      "https://exemplo.com/imagem.jpg",
      "Confira esta imagem!",
    );
    console.log("Imagem enviada com sucesso:", JSON.stringify(result, null, 2));
  } catch (error) {
    console.error("Falha ao enviar imagem:", error);
    process.exit(1);
  }
}

main();
