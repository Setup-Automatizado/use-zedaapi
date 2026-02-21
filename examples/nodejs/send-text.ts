/**
 * Zé da API — Enviar mensagem de texto
 * Documentação: https://api.zedaapi.com/docs
 *
 * Uso: npx tsx send-text.ts
 */

const BASE_URL = process.env.ZEDAAPI_URL || "https://sua-instancia.zedaapi.com";
const CLIENT_TOKEN = process.env.ZEDAAPI_TOKEN || "seu-token-aqui";

interface SendTextResponse {
  key?: {
    remoteJid: string;
    fromMe: boolean;
    id: string;
  };
  message?: Record<string, unknown>;
  error?: string;
}

async function sendText(
  phone: string,
  message: string,
): Promise<SendTextResponse> {
  const response = await fetch(`${BASE_URL}/send-text`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Client-Token": CLIENT_TOKEN,
    },
    body: JSON.stringify({ phone, message }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`Erro ${response.status}: ${error}`);
  }

  return response.json() as Promise<SendTextResponse>;
}

// Exemplo de uso
async function main() {
  try {
    const result = await sendText(
      "5511999999999",
      "Olá! Mensagem enviada via Node.js.",
    );
    console.log(
      "Mensagem enviada com sucesso:",
      JSON.stringify(result, null, 2),
    );
  } catch (error) {
    console.error("Falha ao enviar mensagem:", error);
    process.exit(1);
  }
}

main();
