"use client";

import { useCallback, useMemo, useRef } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import {
	Bold,
	Italic,
	Heading2,
	Heading3,
	Link,
	List,
	Code,
	Image,
	Film,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface ContentEditorProps {
	value: string;
	onChange: (value: string) => void;
	onInsertMedia?: () => void;
	className?: string;
	minRows?: number;
}

interface ToolbarAction {
	icon: React.ElementType;
	label: string;
	prefix: string;
	suffix: string;
	block?: boolean;
}

const TOOLBAR_ACTIONS: ToolbarAction[] = [
	{ icon: Bold, label: "Negrito", prefix: "**", suffix: "**" },
	{ icon: Italic, label: "Italico", prefix: "*", suffix: "*" },
	{
		icon: Heading2,
		label: "Titulo 2",
		prefix: "## ",
		suffix: "",
		block: true,
	},
	{
		icon: Heading3,
		label: "Titulo 3",
		prefix: "### ",
		suffix: "",
		block: true,
	},
	{ icon: Link, label: "Link", prefix: "[", suffix: "](url)" },
	{ icon: List, label: "Lista", prefix: "- ", suffix: "", block: true },
	{
		icon: Code,
		label: "Codigo",
		prefix: "```\n",
		suffix: "\n```",
		block: true,
	},
	{ icon: Image, label: "Imagem", prefix: "![alt](", suffix: ")" },
];

function markdownToHtml(md: string): string {
	let html = md;

	// Code blocks (must be first to avoid inner processing)
	html = html.replace(
		/```(\w*)\n([\s\S]*?)```/g,
		(_match, _lang, code) =>
			`<pre class="rounded-lg bg-muted p-4 overflow-x-auto text-sm"><code>${escapeHtml(code.trim())}</code></pre>`,
	);

	// Inline code
	html = html.replace(
		/`([^`]+)`/g,
		'<code class="rounded bg-muted px-1.5 py-0.5 text-sm">$1</code>',
	);

	// Headings
	html = html.replace(
		/^### (.+)$/gm,
		'<h3 class="text-lg font-semibold mt-4 mb-2">$1</h3>',
	);
	html = html.replace(
		/^## (.+)$/gm,
		'<h2 class="text-xl font-semibold mt-6 mb-3">$1</h2>',
	);
	html = html.replace(
		/^# (.+)$/gm,
		'<h1 class="text-2xl font-bold mt-6 mb-3">$1</h1>',
	);

	// Images
	html = html.replace(
		/!\[([^\]]*)\]\(([^)]+)\)/g,
		'<img src="$2" alt="$1" class="rounded-lg max-w-full my-4" />',
	);

	// Links
	html = html.replace(
		/\[([^\]]+)\]\(([^)]+)\)/g,
		'<a href="$2" class="text-primary underline underline-offset-4" target="_blank" rel="noopener noreferrer">$1</a>',
	);

	// Bold
	html = html.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");

	// Italic
	html = html.replace(/\*([^*]+)\*/g, "<em>$1</em>");

	// Unordered lists
	html = html.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc">$1</li>');
	html = html.replace(
		/(<li[^>]*>.*<\/li>\n?)+/g,
		(match) => `<ul class="my-2 space-y-1">${match}</ul>`,
	);

	// Horizontal rule
	html = html.replace(/^---$/gm, '<hr class="my-6 border-border" />');

	// Paragraphs (consecutive non-empty lines)
	html = html.replace(/^(?!<[a-z])((?:(?!<[a-z]).+\n?)+)/gm, (match) => {
		const trimmed = match.trim();
		if (!trimmed) return "";
		return `<p class="mb-3 leading-relaxed">${trimmed}</p>`;
	});

	// Clean up excess newlines inside tags
	html = html.replace(/\n{3,}/g, "\n\n");

	return html;
}

function escapeHtml(text: string): string {
	return text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;");
}

function countWords(text: string): number {
	const trimmed = text.trim();
	if (!trimmed) return 0;
	return trimmed.split(/\s+/).length;
}

export function ContentEditor({
	value,
	onChange,
	onInsertMedia,
	className,
	minRows = 16,
}: ContentEditorProps) {
	const textareaRef = useRef<HTMLTextAreaElement>(null);

	const insertMarkdown = useCallback(
		(action: ToolbarAction) => {
			const textarea = textareaRef.current;
			if (!textarea) return;

			const start = textarea.selectionStart;
			const end = textarea.selectionEnd;
			const selected = value.substring(start, end);
			const before = value.substring(0, start);
			const after = value.substring(end);

			let newText: string;
			let cursorPos: number;

			if (action.block && start === end) {
				// For block-level items with no selection, ensure new line
				const needsNewline =
					before.length > 0 && !before.endsWith("\n");
				const linePrefix = needsNewline ? "\n" : "";
				newText = `${before}${linePrefix}${action.prefix}${selected}${action.suffix}${after}`;
				cursorPos =
					before.length +
					linePrefix.length +
					action.prefix.length +
					selected.length;
			} else {
				newText = `${before}${action.prefix}${selected || "texto"}${action.suffix}${after}`;
				cursorPos =
					before.length +
					action.prefix.length +
					(selected || "texto").length;
			}

			onChange(newText);

			requestAnimationFrame(() => {
				textarea.focus();
				textarea.setSelectionRange(cursorPos, cursorPos);
			});
		},
		[value, onChange],
	);

	const wordCount = useMemo(() => countWords(value), [value]);
	const preview = useMemo(() => markdownToHtml(value), [value]);

	return (
		<div className={cn("space-y-1.5", className)}>
			<Tabs defaultValue="write">
				<div className="flex items-center justify-between">
					<TabsList>
						<TabsTrigger value="write">Escrever</TabsTrigger>
						<TabsTrigger value="preview">Visualizar</TabsTrigger>
					</TabsList>
				</div>

				<TabsContent value="write" className="space-y-2">
					<div className="flex flex-wrap items-center gap-0.5 rounded-xl border border-border bg-muted/30 p-1">
						{TOOLBAR_ACTIONS.map((action) => (
							<Button
								key={action.label}
								type="button"
								variant="ghost"
								size="icon-xs"
								onClick={() => insertMarkdown(action)}
								aria-label={action.label}
								title={action.label}
							>
								<action.icon className="size-3.5" />
							</Button>
						))}
						{onInsertMedia && (
							<>
								<div className="mx-1 h-4 w-px bg-border" />
								<Button
									type="button"
									variant="ghost"
									size="icon-xs"
									onClick={onInsertMedia}
									aria-label="Inserir midia"
									title="Inserir midia"
								>
									<Film className="size-3.5" />
								</Button>
							</>
						)}
					</div>

					<Textarea
						ref={textareaRef}
						value={value}
						onChange={(e) => onChange(e.target.value)}
						rows={minRows}
						className="min-h-0 font-mono text-sm"
						placeholder="Escreva seu conteudo em Markdown..."
					/>

					<div className="flex justify-end">
						<span className="text-xs tabular-nums text-muted-foreground">
							{wordCount}{" "}
							{wordCount === 1 ? "palavra" : "palavras"}
						</span>
					</div>
				</TabsContent>

				<TabsContent value="preview">
					<div
						className="min-h-[200px] rounded-xl border border-border bg-card p-6 text-sm leading-relaxed"
						// biome-ignore lint/security/noDangerouslySetInnerHtml: markdown preview rendering
						dangerouslySetInnerHTML={{
							__html:
								preview ||
								'<p class="text-muted-foreground">Nenhum conteudo para visualizar</p>',
						}}
					/>
				</TabsContent>
			</Tabs>
		</div>
	);
}
