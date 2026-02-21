"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Check, Copy, Linkedin, MessageCircle } from "lucide-react";

interface ShareButtonsProps {
	url: string;
	title: string;
}

export function ShareButtons({ url, title }: ShareButtonsProps) {
	const [copied, setCopied] = useState(false);

	const encodedUrl = encodeURIComponent(url);
	const encodedTitle = encodeURIComponent(title);

	function copyLink() {
		navigator.clipboard.writeText(url);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	}

	return (
		<div className="flex items-center gap-2">
			<span className="text-sm font-medium text-muted-foreground">
				Compartilhar:
			</span>
			<Button
				variant="outline"
				size="icon-sm"
				asChild
				className="text-muted-foreground hover:text-[#25D366]"
			>
				<a
					href={`https://wa.me/?text=${encodedTitle}%20${encodedUrl}`}
					target="_blank"
					rel="noopener noreferrer"
					aria-label="Compartilhar no WhatsApp"
				>
					<MessageCircle className="size-4" />
				</a>
			</Button>
			<Button
				variant="outline"
				size="icon-sm"
				asChild
				className="text-muted-foreground hover:text-foreground"
			>
				<a
					href={`https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}`}
					target="_blank"
					rel="noopener noreferrer"
					aria-label="Compartilhar no X"
				>
					<svg
						viewBox="0 0 24 24"
						className="size-4 fill-current"
						aria-hidden="true"
					>
						<path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
					</svg>
				</a>
			</Button>
			<Button
				variant="outline"
				size="icon-sm"
				asChild
				className="text-muted-foreground hover:text-[#0077B5]"
			>
				<a
					href={`https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`}
					target="_blank"
					rel="noopener noreferrer"
					aria-label="Compartilhar no LinkedIn"
				>
					<Linkedin className="size-4" />
				</a>
			</Button>
			<Button
				variant="outline"
				size="icon-sm"
				onClick={copyLink}
				className="text-muted-foreground"
				aria-label="Copiar link"
			>
				{copied ? (
					<Check className="size-4 text-green-500" />
				) : (
					<Copy className="size-4" />
				)}
			</Button>
		</div>
	);
}
