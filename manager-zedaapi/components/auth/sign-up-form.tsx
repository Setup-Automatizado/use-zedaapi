"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { signUp } from "@/lib/auth-client";
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

const signUpSchema = z
	.object({
		name: z.string().min(2, "Nome deve ter pelo menos 2 caracteres"),
		email: z.string().email("E-mail inválido"),
		password: z.string().min(8, "Senha deve ter pelo menos 8 caracteres"),
		confirmPassword: z.string(),
		terms: z.literal(true, {
			errorMap: () => ({ message: "Aceite os termos para continuar" }),
		}),
	})
	.refine((data) => data.password === data.confirmPassword, {
		message: "As senhas não conferem",
		path: ["confirmPassword"],
	});

type SignUpValues = z.infer<typeof signUpSchema>;

export function SignUpForm() {
	const router = useRouter();
	const [error, setError] = useState<string | null>(null);

	const {
		register,
		handleSubmit,
		formState: { errors, isSubmitting },
	} = useForm<SignUpValues>({
		resolver: zodResolver(signUpSchema),
		defaultValues: {
			name: "",
			email: "",
			password: "",
			confirmPassword: "",
			terms: undefined,
		},
	});

	async function onSubmit(values: SignUpValues) {
		setError(null);

		const result = await signUp.email({
			name: values.name,
			email: values.email,
			password: values.password,
		});

		if (result.error) {
			setError(
				result.error.message ?? "Erro ao criar conta. Tente novamente.",
			);
			return;
		}

		router.push("/verify-email?registered=true");
	}

	return (
		<Card className="w-full max-w-md">
			<CardHeader className="space-y-1">
				<CardTitle className="text-2xl font-bold tracking-tight">
					Criar conta
				</CardTitle>
				<CardDescription>
					Preencha os dados abaixo para criar sua conta
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
						<FieldLabel htmlFor="name">Nome</FieldLabel>
						<FieldContent>
							<Input
								id="name"
								type="text"
								placeholder="Seu nome completo"
								autoComplete="name"
								disabled={isSubmitting}
								{...register("name")}
							/>
							{errors.name && (
								<FieldError>{errors.name.message}</FieldError>
							)}
						</FieldContent>
					</Field>

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

					<Field>
						<FieldLabel htmlFor="password">Senha</FieldLabel>
						<FieldContent>
							<Input
								id="password"
								type="password"
								placeholder="Min. 8 caracteres"
								autoComplete="new-password"
								disabled={isSubmitting}
								{...register("password")}
							/>
							{errors.password && (
								<FieldError>
									{errors.password.message}
								</FieldError>
							)}
						</FieldContent>
					</Field>

					<Field>
						<FieldLabel htmlFor="confirmPassword">
							Confirmar senha
						</FieldLabel>
						<FieldContent>
							<Input
								id="confirmPassword"
								type="password"
								placeholder="Repita a senha"
								autoComplete="new-password"
								disabled={isSubmitting}
								{...register("confirmPassword")}
							/>
							{errors.confirmPassword && (
								<FieldError>
									{errors.confirmPassword.message}
								</FieldError>
							)}
						</FieldContent>
					</Field>

					<label className="flex items-start gap-2 text-sm">
						<input
							type="checkbox"
							className="mt-0.5 size-4 rounded border-input accent-primary"
							disabled={isSubmitting}
							{...register("terms")}
						/>
						<span className="text-muted-foreground leading-snug">
							Li e aceito os{" "}
							<Link
								href="/terms"
								className="text-foreground underline underline-offset-4"
								target="_blank"
							>
								termos de uso
							</Link>{" "}
							e a{" "}
							<Link
								href="/privacy"
								className="text-foreground underline underline-offset-4"
								target="_blank"
							>
								política de privacidade
							</Link>
						</span>
					</label>
					{errors.terms && (
						<FieldError>{errors.terms.message}</FieldError>
					)}
				</CardContent>

				<CardFooter className="flex flex-col gap-3">
					<Button
						type="submit"
						className="w-full"
						size="lg"
						disabled={isSubmitting}
					>
						{isSubmitting && <Spinner className="mr-2 size-4" />}
						Criar conta
					</Button>
					<p className="text-sm text-muted-foreground">
						Já tem uma conta?{" "}
						<Link
							href="/sign-in"
							className="text-foreground font-medium hover:underline underline-offset-4"
						>
							Entrar
						</Link>
					</p>
				</CardFooter>
			</form>
		</Card>
	);
}
