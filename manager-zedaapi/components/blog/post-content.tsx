interface PostContentProps {
	content: string;
}

export function PostContent({ content }: PostContentProps) {
	const html = markdownToHtml(content);

	return (
		<div
			className="prose prose-zinc dark:prose-invert prose-headings:scroll-mt-20 prose-headings:font-semibold prose-h2:text-2xl prose-h3:text-xl prose-p:leading-relaxed prose-a:text-primary prose-a:no-underline hover:prose-a:underline prose-img:rounded-xl prose-pre:rounded-xl prose-pre:bg-muted max-w-none"
			dangerouslySetInnerHTML={{ __html: html }}
		/>
	);
}

function markdownToHtml(md: string): string {
	let html = md;

	// Code blocks (must be before inline code)
	html = html.replace(
		/```(\w*)\n([\s\S]*?)```/g,
		(_match, lang: string, code: string) => {
			const escaped = escapeHtml(code.trim());
			return `<pre><code class="language-${lang}">${escaped}</code></pre>`;
		},
	);

	// Inline code
	html = html.replace(/`([^`]+)`/g, "<code>$1</code>");

	// Headings (## and ###) with IDs for TOC anchoring
	html = html.replace(/^### (.+)$/gm, (_match, text: string) => {
		const id = slugifyHeading(text);
		return `<h3 id="${id}">${text}</h3>`;
	});
	html = html.replace(/^## (.+)$/gm, (_match, text: string) => {
		const id = slugifyHeading(text);
		return `<h2 id="${id}">${text}</h2>`;
	});
	html = html.replace(/^# (.+)$/gm, "<h1>$1</h1>");

	// Bold and italic
	html = html.replace(/\*\*\*(.+?)\*\*\*/g, "<strong><em>$1</em></strong>");
	html = html.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
	html = html.replace(/\*(.+?)\*/g, "<em>$1</em>");

	// Images
	html = html.replace(
		/!\[([^\]]*)\]\(([^)]+)\)/g,
		'<img src="$2" alt="$1" loading="lazy" />',
	);

	// Links
	html = html.replace(
		/\[([^\]]+)\]\(([^)]+)\)/g,
		'<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>',
	);

	// Unordered lists
	html = html.replace(/^- (.+)$/gm, "<li>$1</li>");
	html = html.replace(/((?:<li>.*<\/li>\n?)+)/g, "<ul>$1</ul>");

	// Ordered lists
	html = html.replace(/^\d+\. (.+)$/gm, "<li>$1</li>");

	// Blockquotes
	html = html.replace(/^> (.+)$/gm, "<blockquote><p>$1</p></blockquote>");

	// Horizontal rules
	html = html.replace(/^---$/gm, "<hr />");

	// Paragraphs: wrap remaining non-tag lines
	html = html
		.split("\n\n")
		.map((block) => {
			const trimmed = block.trim();
			if (!trimmed) return "";
			if (
				trimmed.startsWith("<h") ||
				trimmed.startsWith("<ul") ||
				trimmed.startsWith("<ol") ||
				trimmed.startsWith("<pre") ||
				trimmed.startsWith("<blockquote") ||
				trimmed.startsWith("<hr") ||
				trimmed.startsWith("<img") ||
				trimmed.startsWith("<figure")
			) {
				return trimmed;
			}
			return `<p>${trimmed.replace(/\n/g, "<br />")}</p>`;
		})
		.join("\n");

	return html;
}

function escapeHtml(text: string): string {
	return text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;");
}

function slugifyHeading(text: string): string {
	return text
		.normalize("NFD")
		.replace(/[\u0300-\u036f]/g, "")
		.toLowerCase()
		.replace(/[^a-z0-9\s-]/g, "")
		.replace(/\s+/g, "-");
}
