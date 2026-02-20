import Link from "next/link";
import { Button } from "@/components/ui/button";
import { FileQuestion } from "lucide-react";

export default function NotFound() {
	return (
		<div className="flex min-h-svh flex-col items-center justify-center gap-6 p-4">
			<div className="flex size-16 items-center justify-center rounded-2xl bg-muted">
				<FileQuestion className="size-8 text-muted-foreground" />
			</div>
			<div className="text-center">
				<h1 className="text-4xl font-bold tracking-tight">404</h1>
				<p className="mt-2 text-sm text-muted-foreground">
					A pagina que voce esta procurando nao existe ou foi movida.
				</p>
			</div>
			<Button asChild>
				<Link href="/dashboard">Voltar ao inicio</Link>
			</Button>
		</div>
	);
}
