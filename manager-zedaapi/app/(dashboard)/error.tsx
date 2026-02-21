"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";
import { createClientLogger } from "@/lib/client-logger";

const log = createClientLogger("error:dashboard");

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
		<div className="flex flex-col items-center justify-center gap-6 py-20">
			<div className="flex size-16 items-center justify-center rounded-2xl bg-destructive/10">
				<AlertTriangle className="size-8 text-destructive" />
			</div>
			<div className="text-center">
				<h2 className="text-xl font-bold tracking-tight">
					Algo deu errado
				</h2>
				<p className="mt-2 max-w-sm text-sm text-muted-foreground">
					Ocorreu um erro inesperado. Tente novamente.
				</p>
			</div>
			<Button onClick={reset}>Tentar novamente</Button>
		</div>
	);
}
