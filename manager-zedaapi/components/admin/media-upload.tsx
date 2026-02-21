"use client";

import { useCallback, useRef, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Upload, Link, Youtube, Loader2, Check, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface UploadedMedia {
	url: string;
	s3Key: string;
	type: string;
	filename: string;
}

interface MediaUploadProps {
	onUpload: (media: UploadedMedia) => void;
	accept?: string;
	className?: string;
}

type UploadState = "idle" | "dragging" | "uploading" | "done";

function extractYoutubeId(url: string): string | null {
	const patterns = [
		/(?:youtube\.com\/watch\?v=)([a-zA-Z0-9_-]{11})/,
		/(?:youtu\.be\/)([a-zA-Z0-9_-]{11})/,
		/(?:youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})/,
		/(?:youtube\.com\/shorts\/)([a-zA-Z0-9_-]{11})/,
	];
	for (const pattern of patterns) {
		const match = url.match(pattern);
		if (match?.[1]) return match[1];
	}
	return null;
}

export function MediaUpload({
	onUpload,
	accept = "image/*,video/*,audio/*,.pdf,.doc,.docx",
	className,
}: MediaUploadProps) {
	const [state, setState] = useState<UploadState>("idle");
	const [uploadedFile, setUploadedFile] = useState<UploadedMedia | null>(
		null,
	);
	const [externalUrl, setExternalUrl] = useState("");
	const [error, setError] = useState<string | null>(null);
	const fileInputRef = useRef<HTMLInputElement>(null);

	const resetUpload = useCallback(() => {
		setState("idle");
		setUploadedFile(null);
		setError(null);
		if (fileInputRef.current) {
			fileInputRef.current.value = "";
		}
	}, []);

	const handleFileUpload = useCallback(
		async (file: File) => {
			setState("uploading");
			setError(null);

			try {
				const formData = new FormData();
				formData.append("file", file);

				const response = await fetch("/api/admin/upload", {
					method: "POST",
					body: formData,
				});

				if (!response.ok) {
					const body = (await response.json().catch(() => null)) as {
						error?: string;
					} | null;
					throw new Error(body?.error ?? "Falha no upload");
				}

				const data = (await response.json()) as UploadedMedia;
				setUploadedFile(data);
				setState("done");
				onUpload(data);
			} catch (err) {
				setState("idle");
				setError(
					err instanceof Error ? err.message : "Erro desconhecido",
				);
			}
		},
		[onUpload],
	);

	const handleDrop = useCallback(
		(e: React.DragEvent) => {
			e.preventDefault();
			setState("idle");

			const file = e.dataTransfer.files[0];
			if (file) {
				void handleFileUpload(file);
			}
		},
		[handleFileUpload],
	);

	const handleDragOver = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		setState("dragging");
	}, []);

	const handleDragLeave = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		setState("idle");
	}, []);

	const handleFileSelect = useCallback(
		(e: React.ChangeEvent<HTMLInputElement>) => {
			const file = e.target.files?.[0];
			if (file) {
				void handleFileUpload(file);
			}
		},
		[handleFileUpload],
	);

	const handleExternalUrl = useCallback(() => {
		if (!externalUrl.trim()) return;

		const youtubeId = extractYoutubeId(externalUrl);
		const media: UploadedMedia = {
			url: externalUrl,
			s3Key: "",
			type: youtubeId ? "video/youtube" : "external",
			filename: youtubeId
				? `YouTube: ${youtubeId}`
				: (externalUrl.split("/").pop() ?? "external"),
		};

		setUploadedFile(media);
		setState("done");
		onUpload(media);
	}, [externalUrl, onUpload]);

	const youtubeId = externalUrl ? extractYoutubeId(externalUrl) : null;

	return (
		<div className={cn("space-y-1.5", className)}>
			<Tabs defaultValue="file">
				<TabsList>
					<TabsTrigger value="file">
						<Upload className="mr-1 size-3.5" />
						Arquivo
					</TabsTrigger>
					<TabsTrigger value="url">
						<Link className="mr-1 size-3.5" />
						URL
					</TabsTrigger>
				</TabsList>

				<TabsContent value="file">
					{state === "done" && uploadedFile ? (
						<Card size="sm">
							<CardContent>
								<div className="flex items-center gap-3">
									{uploadedFile.type.startsWith("image/") ? (
										<img
											src={uploadedFile.url}
											alt={uploadedFile.filename}
											className="size-16 rounded-lg border object-cover"
										/>
									) : (
										<div className="flex size-16 items-center justify-center rounded-lg border bg-muted">
											<Check className="size-6 text-primary" />
										</div>
									)}
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium">
											{uploadedFile.filename}
										</p>
										<p className="text-xs text-muted-foreground">
											Upload concluido
										</p>
									</div>
									<Button
										type="button"
										variant="ghost"
										size="icon-xs"
										onClick={resetUpload}
										aria-label="Remover"
									>
										<X className="size-3.5" />
									</Button>
								</div>
							</CardContent>
						</Card>
					) : (
						<div
							onDrop={handleDrop}
							onDragOver={handleDragOver}
							onDragLeave={handleDragLeave}
							onClick={() => fileInputRef.current?.click()}
							onKeyDown={(e) => {
								if (e.key === "Enter" || e.key === " ") {
									fileInputRef.current?.click();
								}
							}}
							role="button"
							tabIndex={0}
							className={cn(
								"flex min-h-[140px] cursor-pointer flex-col items-center justify-center gap-3 rounded-xl border-2 border-dashed transition-colors",
								state === "dragging"
									? "border-primary bg-primary/5"
									: "border-border hover:border-primary/50 hover:bg-muted/30",
								state === "uploading" &&
									"pointer-events-none opacity-60",
							)}
						>
							{state === "uploading" ? (
								<>
									<Loader2 className="size-8 animate-spin text-primary" />
									<p className="text-sm text-muted-foreground">
										Enviando arquivo...
									</p>
								</>
							) : (
								<>
									<Upload className="size-8 text-muted-foreground" />
									<div className="text-center">
										<p className="text-sm font-medium">
											Arraste ou clique para enviar
										</p>
										<p className="mt-1 text-xs text-muted-foreground">
											Imagens, videos, PDFs e documentos
										</p>
									</div>
								</>
							)}
							<input
								ref={fileInputRef}
								type="file"
								accept={accept}
								onChange={handleFileSelect}
								className="hidden"
								aria-label="Selecionar arquivo"
							/>
						</div>
					)}

					{error && (
						<p className="mt-2 text-sm text-destructive">{error}</p>
					)}
				</TabsContent>

				<TabsContent value="url" className="space-y-3">
					<div className="space-y-1.5">
						<Label>URL da midia</Label>
						<div className="flex gap-2">
							<Input
								value={externalUrl}
								onChange={(e) => {
									setExternalUrl(e.target.value);
									setState("idle");
									setUploadedFile(null);
								}}
								placeholder="https://exemplo.com/imagem.jpg ou YouTube URL"
							/>
							<Button
								type="button"
								variant="outline"
								onClick={handleExternalUrl}
								disabled={!externalUrl.trim()}
							>
								Adicionar
							</Button>
						</div>
					</div>

					{youtubeId && (
						<Card size="sm">
							<CardContent>
								<div className="space-y-2">
									<div className="flex items-center gap-2 text-sm font-medium">
										<Youtube className="size-4 text-red-500" />
										Video do YouTube detectado
									</div>
									<div className="aspect-video w-full overflow-hidden rounded-lg">
										<iframe
											src={`https://www.youtube-nocookie.com/embed/${youtubeId}`}
											title="YouTube video preview"
											allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
											allowFullScreen
											className="size-full border-0"
										/>
									</div>
								</div>
							</CardContent>
						</Card>
					)}

					{state === "done" && uploadedFile && !youtubeId && (
						<Card size="sm">
							<CardContent>
								<div className="flex items-center gap-3">
									<Check className="size-5 text-primary" />
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium">
											{uploadedFile.filename}
										</p>
										<p className="truncate text-xs text-muted-foreground">
											{uploadedFile.url}
										</p>
									</div>
									<Button
										type="button"
										variant="ghost"
										size="icon-xs"
										onClick={() => {
											setExternalUrl("");
											setUploadedFile(null);
											setState("idle");
										}}
										aria-label="Remover"
									>
										<X className="size-3.5" />
									</Button>
								</div>
							</CardContent>
						</Card>
					)}
				</TabsContent>
			</Tabs>
		</div>
	);
}
