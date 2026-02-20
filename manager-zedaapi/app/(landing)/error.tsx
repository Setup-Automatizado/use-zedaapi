"use client";

import { useEffect } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";

export default function Error({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	useEffect(() => {
		console.error(error);
	}, [error]);

	return (
		<div className="flex flex-col items-center justify-center gap-6 py-20">
			<div className="flex size-16 items-center justify-center rounded-2xl bg-destructive/10">
				<AlertTriangle className="size-8 text-destructive" />
			</div>
			<div className="text-center">
				<h2 className="text-xl font-bold tracking-tight">
					Algo deu errado
				</h2>
				<p className="mt-2 max-w-sm text-sm text-muted-foreground">
					Ocorreu um erro inesperado. Tente novamente ou entre em
					contato com o suporte se o problema persistir.
				</p>
			</div>
			<div className="flex gap-3">
				<Button onClick={reset}>Tentar novamente</Button>
				<Button variant="outline" asChild>
					<Link href="/">Voltar ao início</Link>
				</Button>
			</div>
		</div>
	);
}
