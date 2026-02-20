"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { Loader2, ArrowLeft, MailCheck } from "lucide-react";

import { authClient } from "@/lib/auth-client";
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

const forgotPasswordSchema = z.object({
	email: z.string().email("E-mail invalido"),
});

type ForgotPasswordValues = z.infer<typeof forgotPasswordSchema>;

export function ForgotPasswordForm() {
	const [sent, setSent] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const {
		register,
		handleSubmit,
		formState: { errors, isSubmitting },
	} = useForm<ForgotPasswordValues>({
		resolver: zodResolver(forgotPasswordSchema),
		defaultValues: { email: "" },
	});

	async function onSubmit(values: ForgotPasswordValues) {
		setError(null);

		const result = await authClient.requestPasswordReset({
			email: values.email,
			redirectTo: "/reset-password",
		});

		if (result.error) {
			setError(
				result.error.message ??
					"Erro ao enviar e-mail. Tente novamente.",
			);
			return;
		}

		setSent(true);
	}

	if (sent) {
		return (
			<Card className="w-full max-w-md">
				<CardHeader className="items-center text-center space-y-1">
					<div className="mb-2 flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
						<MailCheck className="size-6" />
					</div>
					<CardTitle>E-mail enviado</CardTitle>
					<CardDescription>
						Se esse e-mail estiver cadastrado, voce recebera um link
						para redefinir sua senha.
					</CardDescription>
				</CardHeader>
				<CardFooter className="justify-center">
					<Button variant="ghost" size="sm" asChild>
						<Link href="/sign-in">
							<ArrowLeft className="mr-2 size-4" />
							Voltar para login
						</Link>
					</Button>
				</CardFooter>
			</Card>
		);
	}

	return (
		<Card className="w-full max-w-md">
			<CardHeader className="space-y-1">
				<CardTitle className="text-2xl font-bold">Esqueceu a senha?</CardTitle>
				<CardDescription>
					Informe seu e-mail para receber o link de recuperacao
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
						{isSubmitting && (
							<Loader2 className="mr-2 size-4 animate-spin" />
						)}
						Enviar link de recuperacao
					</Button>
					<Button variant="ghost" size="sm" asChild>
						<Link href="/sign-in">
							<ArrowLeft className="mr-2 size-4" />
							Voltar para login
						</Link>
					</Button>
				</CardFooter>
			</form>
		</Card>
	);
}
