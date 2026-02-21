import crypto from "node:crypto";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { requireAdmin } from "@/lib/auth-server";
import { uploadFile } from "@/lib/services/storage/s3-client";

const IMAGE_EXTENSIONS = new Set(["jpg", "jpeg", "png", "webp", "gif", "svg"]);
const VIDEO_EXTENSIONS = new Set(["mp4", "webm"]);
const AUDIO_EXTENSIONS = new Set(["mp3", "wav", "ogg"]);

const MAX_IMAGE_SIZE = 10 * 1024 * 1024; // 10MB
const MAX_VIDEO_SIZE = 50 * 1024 * 1024; // 50MB
const MAX_AUDIO_SIZE = 10 * 1024 * 1024; // 10MB

function getMediaType(ext: string): "image" | "video" | "audio" | null {
	if (IMAGE_EXTENSIONS.has(ext)) return "image";
	if (VIDEO_EXTENSIONS.has(ext)) return "video";
	if (AUDIO_EXTENSIONS.has(ext)) return "audio";
	return null;
}

function getMaxSize(type: "image" | "video" | "audio"): number {
	if (type === "video") return MAX_VIDEO_SIZE;
	return type === "audio" ? MAX_AUDIO_SIZE : MAX_IMAGE_SIZE;
}

function extractYouTubeId(url: string): string | null {
	try {
		const parsed = new URL(url);
		if (
			parsed.hostname === "www.youtube.com" ||
			parsed.hostname === "youtube.com"
		) {
			return parsed.searchParams.get("v");
		}
		if (parsed.hostname === "youtu.be") {
			return parsed.pathname.slice(1);
		}
	} catch {
		// Not a valid URL
	}
	return null;
}

export async function POST(req: NextRequest) {
	try {
		await requireAdmin();

		const contentType = req.headers.get("content-type") || "";

		// Handle JSON body for YouTube/URL embeds
		if (contentType.includes("application/json")) {
			const body = (await req.json()) as {
				type?: string;
				url?: string;
			};

			if (!body.url || !body.type) {
				return NextResponse.json(
					{ error: "Missing url or type" },
					{ status: 400 },
				);
			}

			if (body.type === "youtube") {
				const videoId = extractYouTubeId(body.url);
				if (!videoId) {
					return NextResponse.json(
						{ error: "Invalid YouTube URL" },
						{ status: 400 },
					);
				}
				return NextResponse.json({
					url: `https://www.youtube.com/watch?v=${videoId}`,
					s3Key: null,
					type: "youtube",
					mimeType: null,
					sizeBytes: null,
					filename: null,
				});
			}

			if (body.type === "url") {
				return NextResponse.json({
					url: body.url,
					s3Key: null,
					type: "url",
					mimeType: null,
					sizeBytes: null,
					filename: null,
				});
			}

			return NextResponse.json(
				{ error: "Invalid type. Use 'youtube' or 'url'" },
				{ status: 400 },
			);
		}

		// Handle multipart file upload
		const formData = await req.formData();
		const file = formData.get("file");

		if (!file || !(file instanceof File)) {
			return NextResponse.json(
				{ error: "No file provided" },
				{ status: 400 },
			);
		}

		const filename = file.name;
		const ext = filename.split(".").pop()?.toLowerCase() || "";
		const mediaType = getMediaType(ext);

		if (!mediaType) {
			return NextResponse.json(
				{
					error: `Unsupported file type: .${ext}. Supported: ${[...IMAGE_EXTENSIONS, ...VIDEO_EXTENSIONS, ...AUDIO_EXTENSIONS].join(", ")}`,
				},
				{ status: 400 },
			);
		}

		const maxSize = getMaxSize(mediaType);
		if (file.size > maxSize) {
			return NextResponse.json(
				{
					error: `File too large. Maximum ${mediaType} size: ${maxSize / (1024 * 1024)}MB`,
				},
				{ status: 400 },
			);
		}

		const now = new Date();
		const year = now.getFullYear();
		const month = String(now.getMonth() + 1).padStart(2, "0");
		const s3Key = `blog/${year}/${month}/${crypto.randomUUID()}-${filename}`;

		const buffer = Buffer.from(await file.arrayBuffer());
		const mimeType = file.type || `${mediaType}/${ext}`;

		const url = await uploadFile(s3Key, buffer, mimeType);

		return NextResponse.json({
			url,
			s3Key,
			type: mediaType,
			mimeType,
			sizeBytes: file.size,
			filename,
		});
	} catch (error) {
		if (error instanceof Error && error.message === "NEXT_REDIRECT") {
			throw error;
		}
		return NextResponse.json({ error: "Upload failed" }, { status: 500 });
	}
}
