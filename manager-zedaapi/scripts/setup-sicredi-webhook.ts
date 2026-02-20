import "dotenv/config";
import {
	registerSicrediWebhook,
	getWebhook,
	deleteWebhook,
} from "../lib/services/sicredi/webhook";

// =============================================================================
// Setup Sicredi Webhook — Registra/consulta/remove webhook no Sicredi
//
// USO:
//   bun run scripts/setup-sicredi-webhook.ts register
//   bun run scripts/setup-sicredi-webhook.ts status
//   bun run scripts/setup-sicredi-webhook.ts delete
//
// REQUISITOS:
//   - Certificados mTLS configurados (SICREDI_CERT_PATH, SICREDI_KEY_PATH)
//   - Basic Auth PIX (SICREDI_PIX_BASIC_AUTH)
//   - Chave PIX (SICREDI_PIX_KEY)
//   - URL da aplicação (BETTER_AUTH_URL)
// =============================================================================

const command = process.argv[2] || "status";
const pixKey = process.env.SICREDI_PIX_KEY;

if (!pixKey) {
	console.error("Erro: SICREDI_PIX_KEY não configurada no .env");
	process.exit(1);
}

async function run() {
	switch (command) {
		case "register": {
			console.log("Registrando webhook Sicredi...\n");
			await registerSicrediWebhook();
			console.log("\nWebhook registrado com sucesso!");
			break;
		}

		case "status": {
			console.log(`Consultando webhook para chave PIX: ${pixKey}\n`);
			try {
				const webhook = await getWebhook(pixKey!);
				console.log("Webhook configurado:");
				console.log(`  URL: ${webhook.webhookUrl}`);
				console.log(`  Chave: ${webhook.chave}`);
				console.log(`  Criação: ${webhook.criacao}`);
			} catch (error) {
				if (error instanceof Error && error.message.includes("404")) {
					console.log(
						"Nenhum webhook registrado para esta chave PIX.",
					);
					console.log(
						'Execute "bun run scripts/setup-sicredi-webhook.ts register" para registrar.',
					);
				} else {
					throw error;
				}
			}
			break;
		}

		case "delete": {
			console.log(`Removendo webhook para chave PIX: ${pixKey}\n`);
			await deleteWebhook(pixKey!);
			console.log("Webhook removido com sucesso!");
			break;
		}

		default: {
			console.log(
				"Uso: bun run scripts/setup-sicredi-webhook.ts <comando>\n",
			);
			console.log("Comandos:");
			console.log("  register   Registrar webhook no Sicredi");
			console.log("  status     Consultar webhook atual");
			console.log("  delete     Remover webhook");
		}
	}
}

run().catch((error) => {
	console.error("Erro:", error instanceof Error ? error.message : error);
	process.exit(1);
});
