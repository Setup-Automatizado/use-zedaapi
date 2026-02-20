import { createHash } from "crypto";

/**
 * Generate a valid Sicredi txid (26 uppercase hex chars) from an ID.
 * Uses SHA-256 hash for deterministic, collision-resistant txids.
 */
export function generateTxid(id: string): string {
	const hash = createHash("sha256").update(id).digest("hex");
	return hash.slice(0, 26).toUpperCase();
}

/**
 * Convert cents (integer) to Sicredi amount string format "100.00"
 */
export function formatSicrediAmount(amountCents: number): string {
	return (amountCents / 100).toFixed(2);
}

/**
 * Convert Sicredi amount string "100.00" to cents (integer)
 */
export function parseSicrediAmount(amountStr: string): number {
	return Math.round(parseFloat(amountStr) * 100);
}

/**
 * Map Sicredi charge status to our internal status.
 */
export function mapSicrediStatus(
	sicrediStatus: string,
): "paid" | "pending" | "failed" | "canceled" {
	switch (sicrediStatus) {
		case "CONCLUIDA":
		case "LIQUIDADO":
			return "paid";
		case "ATIVA":
		case "EM_ABERTO":
			return "pending";
		case "REMOVIDA_PELO_USUARIO_RECEBEDOR":
		case "REMOVIDA_PELO_PSP":
			return "failed";
		case "BAIXADO":
			return "canceled";
		default:
			return "pending";
	}
}

/**
 * Calculate due date for boleto (business days from now).
 * Returns YYYY-MM-DD string.
 */
export function calculateDueDate(businessDays: number = 3): string {
	const date = new Date();
	let added = 0;
	while (added < businessDays) {
		date.setDate(date.getDate() + 1);
		const dayOfWeek = date.getDay();
		if (dayOfWeek !== 0 && dayOfWeek !== 6) {
			added++;
		}
	}
	return date.toISOString().split("T")[0]!; // YYYY-MM-DD
}
