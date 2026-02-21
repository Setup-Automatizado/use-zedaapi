import Link from "next/link";
import { Calendar, Clock, Eye } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface PostCardProps {
	slug: string;
	title: string;
	excerpt: string | null;
	coverImageUrl: string | null;
	categoryName: string | null;
	categorySlug: string | null;
	authorName: string;
	publishedAt: Date | null;
	readingTimeMin: number;
	viewCount: number;
}

export function PostCard({
	slug,
	title,
	excerpt,
	coverImageUrl,
	categoryName,
	categorySlug,
	authorName,
	publishedAt,
	readingTimeMin,
	viewCount,
}: PostCardProps) {
	return (
		<Link href={`/blog/${slug}`} className="group block">
			<article className="overflow-hidden rounded-2xl border border-border bg-card transition-all duration-200 hover:-translate-y-0.5 hover:shadow-lg">
				{coverImageUrl && (
					<div className="aspect-video overflow-hidden bg-muted">
						<img
							src={coverImageUrl}
							alt={title}
							className="size-full object-cover transition-transform duration-300 group-hover:scale-105"
						/>
					</div>
				)}
				{!coverImageUrl && (
					<div className="aspect-video bg-gradient-to-br from-primary/10 via-primary/5 to-transparent" />
				)}
				<div className="p-5">
					{categoryName && (
						<Badge
							variant="secondary"
							className="mb-3 text-xs font-medium"
						>
							{categoryName}
						</Badge>
					)}
					<h3 className="line-clamp-2 text-lg font-semibold leading-snug text-foreground group-hover:text-primary transition-colors duration-150">
						{title}
					</h3>
					{excerpt && (
						<p className="mt-2 line-clamp-2 text-sm leading-relaxed text-muted-foreground">
							{excerpt}
						</p>
					)}
					<div className="mt-4 flex items-center gap-4 text-xs text-muted-foreground">
						<span className="font-medium text-foreground/80">
							{authorName}
						</span>
						{publishedAt && (
							<span className="flex items-center gap-1">
								<Calendar className="size-3" />
								{new Date(publishedAt).toLocaleDateString(
									"pt-BR",
									{
										day: "2-digit",
										month: "short",
										year: "numeric",
									},
								)}
							</span>
						)}
						{readingTimeMin > 0 && (
							<span className="flex items-center gap-1">
								<Clock className="size-3" />
								{readingTimeMin} min
							</span>
						)}
						{viewCount > 0 && (
							<span className="flex items-center gap-1">
								<Eye className="size-3" />
								{viewCount}
							</span>
						)}
					</div>
				</div>
			</article>
		</Link>
	);
}
