/**
 * Zé da API — Enviar imagem
 * Documentação: https://api.zedaapi.com/docs
 *
 * Uso: npx tsx send-image.ts
 */

const BASE_URL = process.env.ZEDAAPI_URL || "https://sua-instancia.zedaapi.com";
const CLIENT_TOKEN = process.env.ZEDAAPI_TOKEN || "seu-token-aqui";

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
    headers: {
      "Content-Type": "application/json",
      "Client-Token": CLIENT_TOKEN,
    },
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
