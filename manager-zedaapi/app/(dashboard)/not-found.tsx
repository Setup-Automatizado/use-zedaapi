import Link from "next/link";
import { Button } from "@/components/ui/button";
import { FileQuestion } from "lucide-react";

export default function NotFound() {
	return (
		<div className="flex flex-col items-center justify-center gap-6 py-20">
			<div className="flex size-16 items-center justify-center rounded-2xl bg-muted">
				<FileQuestion className="size-8 text-muted-foreground" />
			</div>
			<div className="text-center">
				<h2 className="text-xl font-bold tracking-tight">
					Página não encontrada
				</h2>
				<p className="mt-2 max-w-sm text-sm text-muted-foreground">
					A página que você procura não existe ou foi removida.
				</p>
			</div>
			<Button asChild>
				<Link href="/">Voltar ao início</Link>
			</Button>
		</div>
	);
}
