"use client";

import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { MailCheck, ArrowLeft } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
	Card,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export function VerifyEmailNotice() {
	const searchParams = useSearchParams();
	const isRegistered = searchParams.get("registered") === "true";

	return (
		<Card className="w-full max-w-md">
			<CardHeader className="items-center text-center space-y-1">
				<div className="mb-2 flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
					<MailCheck className="size-6" />
				</div>
				<CardTitle className="text-xl font-semibold">
					Verifique seu e-mail
				</CardTitle>
				<CardDescription>
					{isRegistered
						? "Sua conta foi criada. Enviamos um link de verificacao para o seu e-mail. Verifique sua caixa de entrada e spam."
						: "Enviamos um link de verificacao para o seu e-mail. Verifique sua caixa de entrada e spam."}
				</CardDescription>
			</CardHeader>
			<CardFooter className="justify-center">
				<Button variant="ghost" size="sm" asChild>
					<Link href="/login">
						<ArrowLeft className="mr-2 size-4" />
						Ir para login
					</Link>
				</Button>
			</CardFooter>
		</Card>
	);
}
