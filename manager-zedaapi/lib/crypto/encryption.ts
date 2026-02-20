import {
	createCipheriv,
	createDecipheriv,
	createHash,
	randomBytes,
	scrypt,
} from "crypto";
import { promisify } from "util";

const scryptAsync = promisify(scrypt);

const ALGORITHM = "aes-256-gcm";
const IV_LENGTH = 16;
const AUTH_TAG_LENGTH = 16;
const SALT_LENGTH = 32;

function getEncryptionKey(): string {
	const key = process.env.ENCRYPTION_KEY;
	if (!key || key.length !== 64) {
		throw new Error(
			"ENCRYPTION_KEY must be a 64-character hex string (32 bytes)",
		);
	}
	return key;
}

export async function encrypt(plaintext: string): Promise<string> {
	const key = Buffer.from(getEncryptionKey(), "hex");
	const iv = randomBytes(IV_LENGTH);
	const salt = randomBytes(SALT_LENGTH);

	const derivedKey = (await scryptAsync(key, salt, 32)) as Buffer;
	const cipher = createCipheriv(ALGORITHM, derivedKey, iv);

	const encrypted = Buffer.concat([
		cipher.update(plaintext, "utf8"),
		cipher.final(),
	]);

	const authTag = cipher.getAuthTag();

	const combined = Buffer.concat([salt, iv, authTag, encrypted]);

	return combined.toString("base64");
}

export async function decrypt(ciphertext: string): Promise<string> {
	const key = Buffer.from(getEncryptionKey(), "hex");
	const combined = Buffer.from(ciphertext, "base64");

	const salt = combined.subarray(0, SALT_LENGTH);
	const iv = combined.subarray(SALT_LENGTH, SALT_LENGTH + IV_LENGTH);
	const authTag = combined.subarray(
		SALT_LENGTH + IV_LENGTH,
		SALT_LENGTH + IV_LENGTH + AUTH_TAG_LENGTH,
	);
	const encrypted = combined.subarray(
		SALT_LENGTH + IV_LENGTH + AUTH_TAG_LENGTH,
	);

	const derivedKey = (await scryptAsync(key, salt, 32)) as Buffer;
	const decipher = createDecipheriv(ALGORITHM, derivedKey, iv);
	decipher.setAuthTag(authTag);

	const decrypted = Buffer.concat([
		decipher.update(encrypted),
		decipher.final(),
	]);

	return decrypted.toString("utf8");
}

export function generateEncryptionKey(): string {
	return randomBytes(32).toString("hex");
}

export function hashToken(token: string): string {
	return createHash("sha256").update(token).digest("hex");
}
