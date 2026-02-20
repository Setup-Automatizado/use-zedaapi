export default function Loading() {
	return (
		<div className="flex min-h-svh items-center justify-center">
			<div className="flex flex-col items-center gap-3">
				<div className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground font-bold text-lg">
					Z
				</div>
				<div className="size-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
			</div>
		</div>
	);
}
