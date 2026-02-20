interface SectionHeaderProps {
	overline?: string;
	title: string;
	description?: string;
	align?: "left" | "center";
}

export function SectionHeader({
	overline,
	title,
	description,
	align = "center",
}: SectionHeaderProps) {
	const textAlign = align === "center" ? "text-center" : "text-left";
	return (
		<div className={`${textAlign} space-y-3`}>
			{overline && (
				<p className="text-sm font-semibold uppercase tracking-wider text-primary">
					{overline}
				</p>
			)}
			<h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
				{title}
			</h2>
			{description && (
				<p className="mx-auto max-w-2xl text-lg leading-relaxed text-muted-foreground">
					{description}
				</p>
			)}
		</div>
	);
}
