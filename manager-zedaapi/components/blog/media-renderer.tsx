"use client";

interface MediaRendererProps {
	type: string;
	url: string;
	alt?: string | null;
	caption?: string | null;
}

export function MediaRenderer({ type, url, alt, caption }: MediaRendererProps) {
	if (type === "youtube") {
		const videoId = extractYoutubeId(url);
		if (!videoId) return null;

		return (
			<figure className="my-6">
				<div className="aspect-video overflow-hidden rounded-xl">
					<iframe
						src={`https://www.youtube-nocookie.com/embed/${videoId}`}
						title={alt ?? "Video"}
						allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
						allowFullScreen
						className="size-full"
						loading="lazy"
					/>
				</div>
				{caption && (
					<figcaption className="mt-2 text-center text-sm text-muted-foreground">
						{caption}
					</figcaption>
				)}
			</figure>
		);
	}

	if (type === "video") {
		return (
			<figure className="my-6">
				<video
					src={url}
					controls
					preload="metadata"
					className="w-full rounded-xl"
				>
					<track kind="captions" />
				</video>
				{caption && (
					<figcaption className="mt-2 text-center text-sm text-muted-foreground">
						{caption}
					</figcaption>
				)}
			</figure>
		);
	}

	if (type === "audio") {
		return (
			<figure className="my-6">
				<audio src={url} controls preload="metadata" className="w-full">
					<track kind="captions" />
				</audio>
				{caption && (
					<figcaption className="mt-2 text-center text-sm text-muted-foreground">
						{caption}
					</figcaption>
				)}
			</figure>
		);
	}

	return (
		<figure className="my-6">
			<img
				src={url}
				alt={alt ?? ""}
				className="w-full rounded-xl"
				loading="lazy"
			/>
			{caption && (
				<figcaption className="mt-2 text-center text-sm text-muted-foreground">
					{caption}
				</figcaption>
			)}
		</figure>
	);
}

function extractYoutubeId(url: string): string | null {
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
		// not a valid URL
	}
	return null;
}
