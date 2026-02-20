import type { S3Client as BunS3Client } from "bun";

let _s3: BunS3Client | null = null;

function getS3(): BunS3Client {
	if (!_s3) {
		// eslint-disable-next-line @typescript-eslint/no-require-imports
		const { S3Client } = require("bun") as typeof import("bun");
		_s3 = new S3Client({
			accessKeyId: process.env.S3_ACCESS_KEY_ID!,
			secretAccessKey: process.env.S3_SECRET_ACCESS_KEY!,
			bucket: process.env.S3_BUCKET!,
			region: process.env.S3_REGION || "us-east-1",
			endpoint: process.env.S3_ENDPOINT,
			virtualHostedStyle: process.env.S3_PATH_STYLE !== "true",
		});
	}
	return _s3;
}

export async function uploadFile(
	key: string,
	body: Buffer | ArrayBuffer | string | Blob,
	contentType?: string,
): Promise<string> {
	const file = getS3().file(key);
	await file.write(body, { type: contentType });
	return getPublicUrl(key);
}

export async function deleteFile(key: string): Promise<void> {
	const file = getS3().file(key);
	await file.delete();
}

export async function getFile(key: string): Promise<Buffer | null> {
	try {
		const file = getS3().file(key);
		const exists = await file.exists();
		if (!exists) return null;

		const response = await file.arrayBuffer();
		return Buffer.from(response);
	} catch {
		return null;
	}
}

export function getPublicUrl(key: string): string {
	const bucket = process.env.S3_BUCKET!;
	const publicUrl = process.env.S3_PUBLIC_URL;
	const pathStyle = process.env.S3_PATH_STYLE === "true";

	if (publicUrl) {
		return pathStyle
			? `${publicUrl}/${bucket}/${key}`
			: `${publicUrl}/${key}`;
	}

	const region = process.env.S3_REGION || "us-east-1";
	return `https://${bucket}.s3.${region}.amazonaws.com/${key}`;
}

export async function generatePresignedUploadUrl(
	key: string,
	expiresIn = 3600,
): Promise<string> {
	const file = getS3().file(key);
	return await file.presign({ expiresIn, method: "PUT" });
}

export async function generatePresignedDownloadUrl(
	key: string,
	expiresIn = 3600,
): Promise<string> {
	const file = getS3().file(key);
	return await file.presign({ expiresIn, method: "GET" });
}

export function extractKeyFromUrl(url: string): string {
	const bucket = process.env.S3_BUCKET!;
	const publicUrl = process.env.S3_PUBLIC_URL;
	const pathStyle = process.env.S3_PATH_STYLE === "true";

	try {
		const parsed = new URL(url);
		const path = parsed.pathname;

		if (pathStyle && publicUrl) {
			const prefix = `/${bucket}/`;
			if (path.startsWith(prefix)) {
				return decodeURIComponent(path.slice(prefix.length));
			}
		}

		if (path.startsWith("/")) {
			return decodeURIComponent(path.slice(1));
		}

		return decodeURIComponent(path);
	} catch {
		const bucketPrefix = `/${bucket}/`;
		const idx = url.indexOf(bucketPrefix);
		if (idx !== -1) {
			return url.slice(idx + bucketPrefix.length);
		}
		return url;
	}
}

export function isS3Url(url: string): boolean {
	const publicUrl = process.env.S3_PUBLIC_URL;
	const bucket = process.env.S3_BUCKET!;

	if (publicUrl && url.startsWith(publicUrl)) return true;
	if (url.includes(`${bucket}.s3.`)) return true;
	if (url.includes(`/${bucket}/`)) return true;

	return false;
}

const SECURE_PREFIXES = ["invoices/", "certificates/", "documents/"];

export function isSecureKey(key: string): boolean {
	return SECURE_PREFIXES.some((prefix) => key.startsWith(prefix));
}

export { getS3 as s3 };
