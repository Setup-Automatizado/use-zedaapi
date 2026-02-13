/**
 * S3 Client Configuration
 *
 * Uses Bun's native S3 for file uploads.
 * Configured for MinIO/S3-compatible storage.
 *
 * @module lib/s3
 */

// S3 credentials from environment
const s3Credentials = {
	accessKeyId: process.env.S3_ACCESS_KEY || "",
	secretAccessKey: process.env.S3_SECRET_KEY || "",
	endpoint: process.env.S3_ENDPOINT || "",
	bucket: process.env.S3_BUCKET || "",
	region: process.env.S3_REGION || "us-east-1",
};

// Validate S3 configuration
export function isS3Configured(): boolean {
	return !!(
		s3Credentials.accessKeyId &&
		s3Credentials.secretAccessKey &&
		s3Credentials.endpoint &&
		s3Credentials.bucket
	);
}

// Public URL base for uploaded files
export const s3PublicUrl =
	process.env.S3_PUBLIC_URL ||
	`${s3Credentials.endpoint}/${s3Credentials.bucket}`;

/**
 * Get S3 file reference using Bun's native S3
 * Uses the global Bun.s3 API
 */
function getS3File(key: string) {
	// @ts-expect-error - Bun global is available at runtime
	return Bun.s3.file(key, {
		accessKeyId: s3Credentials.accessKeyId,
		secretAccessKey: s3Credentials.secretAccessKey,
		endpoint: s3Credentials.endpoint,
		bucket: s3Credentials.bucket,
		region: s3Credentials.region,
	});
}

/**
 * Upload a file to S3
 *
 * @param key - The file path/name in the bucket
 * @param data - The file data (Buffer, Blob, or string)
 * @param contentType - The MIME type of the file
 * @returns The public URL of the uploaded file
 */
export async function uploadFile(
	key: string,
	data: Buffer | Blob | string,
	contentType: string,
): Promise<string> {
	if (!isS3Configured()) {
		throw new Error("S3 is not configured");
	}

	const s3file = getS3File(key);
	await s3file.write(data, {
		type: contentType,
		acl: "public-read",
	});

	return `${s3PublicUrl}/${key}`;
}

/**
 * Delete a file from S3
 *
 * @param key - The file path/name in the bucket
 */
export async function deleteFile(key: string): Promise<void> {
	if (!isS3Configured()) {
		throw new Error("S3 is not configured");
	}

	const s3file = getS3File(key);
	await s3file.delete();
}

/**
 * Check if a file exists in S3
 *
 * @param key - The file path/name in the bucket
 * @returns True if the file exists
 */
export async function fileExists(key: string): Promise<boolean> {
	if (!isS3Configured()) {
		return false;
	}

	const s3file = getS3File(key);
	return await s3file.exists();
}

/**
 * Generate a unique filename for uploads
 *
 * @param originalName - The original filename
 * @param prefix - Optional prefix for the path
 * @returns A unique filename with path
 */
export function generateUniqueFilename(
	originalName: string,
	prefix: string = "uploads",
): string {
	const timestamp = Date.now();
	const random = Math.random().toString(36).substring(2, 8);
	const ext = originalName.split(".").pop() || "bin";
	return `${prefix}/${timestamp}-${random}.${ext}`;
}

/**
 * Get allowed image extensions
 */
export const ALLOWED_IMAGE_TYPES = [
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
];

/**
 * Maximum file size for avatars (2MB)
 */
export const MAX_AVATAR_SIZE = 2 * 1024 * 1024;

/**
 * Validate image file
 */
export function validateImageFile(file: File): {
	valid: boolean;
	error?: string;
} {
	if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
		return {
			valid: false,
			error: "Invalid file type. Allowed: JPEG, PNG, GIF, WebP",
		};
	}

	if (file.size > MAX_AVATAR_SIZE) {
		return {
			valid: false,
			error: "File too large. Maximum size: 2MB",
		};
	}

	return { valid: true };
}
