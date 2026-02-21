"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { signIn } from "@/lib/auth-client";
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

const signInSchema = z.object({
	email: z.string().email("E-mail inválido"),
	password: z.string().min(1, "Senha obrigatória"),
});

type SignInValues = z.infer<typeof signInSchema>;

export function SignInForm() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const callbackUrl = searchParams.get("callbackUrl") ?? "/painel";
	const [error, setError] = useState<string | null>(null);

	const {
		register,
		handleSubmit,
		formState: { errors, isSubmitting },
	} = useForm<SignInValues>({
		resolver: zodResolver(signInSchema),
		defaultValues: { email: "", password: "" },
	});

	async function onSubmit(values: SignInValues) {
		setError(null);

		const result = await signIn.email({
			email: values.email,
			password: values.password,
		});

		if (result.error) {
			setError(
				result.error.message ?? "Erro ao fazer login. Tente novamente.",
			);
			return;
		}

		router.push(callbackUrl);
		router.refresh();
	}

	return (
		<Card className="w-full max-w-md">
			<CardHeader className="space-y-1">
				<CardTitle className="text-2xl font-bold tracking-tight">
					Entrar
				</CardTitle>
				<CardDescription>
					Entre com seu e-mail e senha para acessar sua conta
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

					<Field>
						<FieldLabel htmlFor="password">Senha</FieldLabel>
						<FieldContent>
							<Input
								id="password"
								type="password"
								placeholder="********"
								autoComplete="current-password"
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

					<div className="flex justify-end">
						<Link
							href="/esqueci-senha"
							className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-150"
						>
							Esqueceu a senha?
						</Link>
					</div>
				</CardContent>

				<CardFooter className="flex flex-col gap-3">
					<Button
						type="submit"
						className="w-full"
						size="lg"
						disabled={isSubmitting}
					>
						{isSubmitting && <Spinner className="mr-2 size-4" />}
						Entrar
					</Button>
					<p className="text-sm text-muted-foreground">
						Não tem uma conta?{" "}
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
