"use client";

import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

interface TocItem {
	id: string;
	text: string;
	level: number;
}

interface TocProps {
	content: string;
}

export function TableOfContents({ content }: TocProps) {
	const [activeId, setActiveId] = useState("");
	const items = extractHeadings(content);

	useEffect(() => {
		if (items.length === 0) return;

		const observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					if (entry.isIntersecting) {
						setActiveId(entry.target.id);
					}
				}
			},
			{ rootMargin: "-80px 0px -80% 0px" },
		);

		for (const item of items) {
			const el = document.getElementById(item.id);
			if (el) observer.observe(el);
		}

		return () => observer.disconnect();
	}, [items]);

	if (items.length < 2) return null;

	return (
		<nav className="space-y-1" aria-label="Indice">
			<p className="mb-3 text-sm font-semibold text-foreground">
				Neste artigo
			</p>
			{items.map((item) => (
				<a
					key={item.id}
					href={`#${item.id}`}
					className={cn(
						"block text-sm leading-relaxed transition-colors duration-150 hover:text-foreground",
						item.level === 3 && "pl-4",
						activeId === item.id
							? "font-medium text-primary"
							: "text-muted-foreground",
					)}
				>
					{item.text}
				</a>
			))}
		</nav>
	);
}

function extractHeadings(content: string): TocItem[] {
	const items: TocItem[] = [];
	const lines = content.split("\n");

	for (const line of lines) {
		const match = line.match(/^(#{2,3})\s+(.+)/);
		if (match) {
			const level = match[1]!.length;
			const text = match[2]!.trim();
			const id = text
				.normalize("NFD")
				.replace(/[\u0300-\u036f]/g, "")
				.toLowerCase()
				.replace(/[^a-z0-9\s-]/g, "")
				.replace(/\s+/g, "-");
			items.push({ id, text, level });
		}
	}

	return items;
}
