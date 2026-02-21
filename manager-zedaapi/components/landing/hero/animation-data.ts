export type MessageType = "text" | "image" | "audio";

export interface DemoScenario {
	id: string;
	type: MessageType;
	command: string;
	response: string;
	message: {
		text?: string;
		caption?: string;
		audioDuration?: string;
	};
	copyCommand: string;
}

const API = "https://api.zedaapi.com";
const PATH = "/instances/{id}/token/{tkn}";

export const DEMO_SCENARIOS: DemoScenario[] = [
	{
		id: "send-text",
		type: "text",
		command: `$ curl -X POST \\
  ${API}${PATH}/send-text \\
  -H "Client-Token: zApi_T0K3N_S3gUR0" \\
  -d '{
    "phone": "5521999999999",
    "message": "Pedido #1234 confirmado!"
  }'`,
		response: `{
  "messageId": "BAE5F4C2A1B3D6E8",
  "status": "sent"
}`,
		message: {
			text: "Pedido #1234 confirmado!",
		},
		copyCommand: `curl -X POST ${API}${PATH}/send-text \\
  -H "Client-Token: YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{ "phone": "5521999999999", "message": "Pedido #1234 confirmado!" }'`,
	},
	{
		id: "send-image",
		type: "image",
		command: `$ curl -X POST \\
  ${API}${PATH}/send-image \\
  -H "Client-Token: zApi_T0K3N_S3gUR0" \\
  -d '{
    "phone": "5521999999999",
    "image": "https://cdn.loja.com/recibo.png",
    "caption": "Comprovante - R$ 197,00"
  }'`,
		response: `{
  "messageId": "CAE7D3B1F2A4C6E9",
  "status": "sent"
}`,
		message: {
			caption: "Comprovante - R$ 197,00",
		},
		copyCommand: `curl -X POST ${API}${PATH}/send-image \\
  -H "Client-Token: YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{ "phone": "5521999999999", "image": "https://cdn.loja.com/recibo.png", "caption": "Comprovante - R$ 197,00" }'`,
	},
	{
		id: "send-audio",
		type: "audio",
		command: `$ curl -X POST \\
  ${API}${PATH}/send-audio \\
  -H "Client-Token: zApi_T0K3N_S3gUR0" \\
  -d '{
    "phone": "5521999999999",
    "audio": "https://cdn.loja.com/confirm.ogg"
  }'`,
		response: `{
  "messageId": "DAF8E2C3A5B7D1F4",
  "status": "sent"
}`,
		message: {
			audioDuration: "0:14",
		},
		copyCommand: `curl -X POST ${API}${PATH}/send-audio \\
  -H "Client-Token: YOUR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{ "phone": "5521999999999", "audio": "https://cdn.loja.com/confirm.ogg" }'`,
	},
];

// Timing constants (ms)
export const TIMING = {
	charDelay: 18,
	responseDelay: 700,
	typingIndicatorDuration: 800,
	messageAppearDelay: 800,
	sentDelay: 1000,
	deliveredDelay: 1200,
	readDelay: 1500,
	pauseBetweenScenarios: 3000,
} as const;
