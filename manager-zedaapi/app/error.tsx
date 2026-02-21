"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";
import { createClientLogger } from "@/lib/client-logger";

const log = createClientLogger("error");

export default function Error({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	useEffect(() => {
		log.error("Unhandled error", {
			message: error.message,
			digest: error.digest,
		});
	}, [error]);

	return (
		<div className="flex min-h-svh flex-col items-center justify-center gap-6 p-4">
			<div className="flex size-16 items-center justify-center rounded-2xl bg-destructive/10">
				<AlertTriangle className="size-8 text-destructive" />
			</div>
			<div className="text-center">
				<h1 className="text-2xl font-bold tracking-tight">
					Algo deu errado
				</h1>
				<p className="mt-2 max-w-sm text-sm text-muted-foreground">
					Ocorreu um erro inesperado. Tente novamente ou entre em
					contato com o suporte se o problema persistir.
				</p>
			</div>
			<div className="flex gap-3">
				<Button onClick={reset}>Tentar novamente</Button>
				<Button variant="outline" asChild>
					<a href="mailto:suporte@zedaapi.com">Reportar problema</a>
				</Button>
			</div>
		</div>
	);
}
