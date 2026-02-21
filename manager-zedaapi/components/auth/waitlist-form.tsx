"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { CheckCircle2, ArrowLeft } from "lucide-react";

import { authClient } from "@/lib/auth-client";
import { Spinner } from "@/components/ui/spinner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Field,
	FieldContent,
	FieldError,
	FieldLabel,
} from "@/components/ui/field";

const waitlistSchema = z.object({
	email: z.string().email("E-mail inválido"),
});

type WaitlistValues = z.infer<typeof waitlistSchema>;

export function WaitlistForm() {
	const [submitted, setSubmitted] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const {
		register,
		handleSubmit,
		formState: { errors, isSubmitting },
	} = useForm<WaitlistValues>({
		resolver: zodResolver(waitlistSchema),
		defaultValues: { email: "" },
	});

	async function onSubmit(values: WaitlistValues) {
		setError(null);

		const result = await authClient.waitlist.join({
			email: values.email,
		});

		if (result.error) {
			setError(
				result.error.message ??
					"Erro ao entrar na lista de espera. Tente novamente.",
			);
			return;
		}

		setSubmitted(true);
	}

	if (submitted) {
		return (
			<Card className="w-full max-w-md">
				<CardHeader className="items-center text-center space-y-1">
					<div className="mb-2 flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
						<CheckCircle2 className="size-6" />
					</div>
					<CardTitle className="text-xl font-semibold">
						Você está na lista!
					</CardTitle>
					<CardDescription>
						Avisaremos por e-mail quando sua conta estiver
						disponível. Fique de olho na sua caixa de entrada.
					</CardDescription>
				</CardHeader>
				<CardFooter className="justify-center">
					<Button variant="ghost" size="sm" asChild>
						<Link href="/">
							<ArrowLeft className="mr-2 size-4" />
							Voltar para início
						</Link>
					</Button>
				</CardFooter>
			</Card>
		);
	}

	return (
		<Card className="w-full max-w-md">
			<CardHeader className="space-y-1">
				<CardTitle className="text-2xl font-bold tracking-tight">
					Lista de espera
				</CardTitle>
				<CardDescription>
					O Zé da API Manager está em acesso antecipado. Cadastre seu
					e-mail para receber um convite.
				</CardDescription>
			</CardHeader>
			<form onSubmit={handleSubmit(onSubmit)}>
				<CardContent className="flex flex-col gap-4">
					{error && (
						<div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
							{error}
						</div>
					)}

					<Field>
						<FieldLabel htmlFor="email">E-mail</FieldLabel>
						<FieldContent>
							<Input
								id="email"
								type="email"
								placeholder="seu@email.com"
								autoComplete="email"
								disabled={isSubmitting}
								{...register("email")}
							/>
							{errors.email && (
								<FieldError>{errors.email.message}</FieldError>
							)}
						</FieldContent>
					</Field>
				</CardContent>

				<CardFooter className="flex flex-col gap-3">
					<Button
						type="submit"
						className="w-full"
						size="lg"
						disabled={isSubmitting}
					>
						{isSubmitting && <Spinner className="mr-2 size-4" />}
						Entrar na lista de espera
					</Button>
					<p className="text-sm text-muted-foreground">
						Já tem um convite?{" "}
						<Link
							href="/cadastro"
							className="text-foreground font-medium hover:underline underline-offset-4"
						>
							Criar conta
						</Link>
					</p>
				</CardFooter>
			</form>
		</Card>
	);
}
