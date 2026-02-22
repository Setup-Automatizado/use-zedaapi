/**
 * Zé da API — Listener de Webhooks
 * Documentação: https://api.zedaapi.com/docs
 *
 * Servidor HTTP simples que recebe eventos do Zé da API via webhook.
 *
 * Uso: npx tsx webhook-listener.ts
 */

import {
  createServer,
  type IncomingMessage,
  type ServerResponse,
} from "node:http";

const PORT = parseInt(process.env.PORT || "3000", 10);

interface WebhookEvent {
  event?: string;
  instanceId?: string;
  data?: Record<string, unknown>;
  [key: string]: unknown;
}

function parseBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    req.on("data", (chunk: Buffer) => chunks.push(chunk));
    req.on("end", () => resolve(Buffer.concat(chunks).toString()));
    req.on("error", reject);
  });
}

const server = createServer(
  async (req: IncomingMessage, res: ServerResponse) => {
    // Health check
    if (req.method === "GET" && req.url === "/health") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ status: "ok" }));
      return;
    }

    // Webhook endpoint
    if (req.method === "POST" && req.url === "/webhook") {
      try {
        const body = await parseBody(req);
        const event: WebhookEvent = JSON.parse(body);

        const eventType = event.event || "unknown";
        const timestamp = new Date().toISOString();

        console.log(`[${timestamp}] Evento recebido: ${eventType}`);
        console.log(JSON.stringify(event, null, 2));
        console.log("---");

        // Processar diferentes tipos de evento
        switch (eventType) {
          case "messages.upsert":
            console.log("Nova mensagem recebida!");
            break;
          case "messages.update":
            console.log("Status de mensagem atualizado!");
            break;
          case "connection.update":
            console.log("Status da conexão alterado!");
            break;
          default:
            console.log(`Evento não tratado: ${eventType}`);
        }

        // Sempre retornar 200 rapidamente para o Zé da API
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ received: true }));
      } catch (error) {
        console.error("Erro ao processar webhook:", error);
        res.writeHead(400, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "Payload inválido" }));
      }
      return;
    }

    // Rota não encontrada
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "Rota não encontrada" }));
  },
);

server.listen(PORT, () => {
  console.log(`Webhook listener rodando em http://localhost:${PORT}/webhook`);
  console.log("Aguardando eventos...");
  console.log("");
  console.log("Configure o webhook no Zé da API apontando para:");
  console.log(`  http://SEU-IP:${PORT}/webhook`);
  console.log("");
});
