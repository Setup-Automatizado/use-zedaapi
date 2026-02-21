/**
 * Zé da API — Enviar mensagem de texto
 * Documentação: https://api.zedaapi.com/docs
 *
 * Uso: npx tsx send-text.ts
 */

// Configuração — substitua pelos seus dados
const HOST = process.env.ZEDAAPI_HOST || "https://sua-instancia.zedaapi.com";
const INSTANCE_ID = process.env.ZEDAAPI_INSTANCE_ID || "sua-instancia";
const INSTANCE_TOKEN = process.env.ZEDAAPI_INSTANCE_TOKEN || "seu-token-aqui";

// URL base com autenticação embutida na rota
const BASE_URL = `${HOST}/instances/${INSTANCE_ID}/token/${INSTANCE_TOKEN}`;

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
    headers: { "Content-Type": "application/json" },
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
