"use client";

import { useState, useRef, useCallback, type KeyboardEvent } from "react";
import { useRouter } from "next/navigation";
import { ShieldCheck } from "lucide-react";

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

const CODE_LENGTH = 6;

export function TwoFactorForm() {
	const router = useRouter();
	const [code, setCode] = useState<string[]>(Array(CODE_LENGTH).fill(""));
	const [error, setError] = useState<string | null>(null);
	const [isSubmitting, setIsSubmitting] = useState(false);
	const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

	const setRef = useCallback(
		(index: number) => (el: HTMLInputElement | null) => {
			inputRefs.current[index] = el;
		},
		[],
	);

	function handleChange(index: number, value: string) {
		if (!/^\d*$/.test(value)) return;

		const newCode = [...code];

		if (value.length > 1) {
			// Paste handling
			const digits = value.replace(/\D/g, "").slice(0, CODE_LENGTH);
			for (let i = 0; i < CODE_LENGTH; i++) {
				newCode[i] = digits[i] ?? "";
			}
			setCode(newCode);
			const nextIndex = Math.min(digits.length, CODE_LENGTH - 1);
			inputRefs.current[nextIndex]?.focus();

			if (digits.length === CODE_LENGTH) {
				submitCode(newCode.join(""));
			}
			return;
		}

		newCode[index] = value;
		setCode(newCode);

		if (value && index < CODE_LENGTH - 1) {
			inputRefs.current[index + 1]?.focus();
		}

		if (
			newCode.every((d) => d !== "") &&
			newCode.join("").length === CODE_LENGTH
		) {
			submitCode(newCode.join(""));
		}
	}

	function handleKeyDown(index: number, e: KeyboardEvent<HTMLInputElement>) {
		if (e.key === "Backspace" && !code[index] && index > 0) {
			inputRefs.current[index - 1]?.focus();
		}
	}

	async function submitCode(totpCode: string) {
		setIsSubmitting(true);
		setError(null);

		const result = await authClient.twoFactor.verifyTotp({
			code: totpCode,
		});

		if (result.error) {
			setError(
				result.error.message ?? "Código inválido. Tente novamente.",
			);
			setCode(Array(CODE_LENGTH).fill(""));
			inputRefs.current[0]?.focus();
			setIsSubmitting(false);
			return;
		}

		router.push("/painel");
		router.refresh();
	}

	return (
		<Card className="w-full max-w-md">
			<CardHeader className="items-center text-center space-y-1">
				<div className="mb-2 flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
					<ShieldCheck className="size-6" />
				</div>
				<CardTitle>Verificação em duas etapas</CardTitle>
				<CardDescription>
					Digite o código de 6 dígitos do seu aplicativo autenticador
				</CardDescription>
			</CardHeader>

			<CardContent className="flex flex-col items-center gap-4">
				{error && (
					<div className="w-full rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive text-center">
						{error}
					</div>
				)}

				<div className="flex gap-2">
					{code.map((digit, index) => (
						<Input
							key={index}
							ref={setRef(index)}
							type="text"
							inputMode="numeric"
							maxLength={CODE_LENGTH}
							value={digit}
							onChange={(e) =>
								handleChange(index, e.target.value)
							}
							onKeyDown={(e) => handleKeyDown(index, e)}
							disabled={isSubmitting}
							className="size-12 text-center text-lg font-mono tabular-nums"
							autoFocus={index === 0}
						/>
					))}
				</div>
			</CardContent>

			<CardFooter className="flex flex-col gap-3">
				<Button
					className="w-full"
					size="lg"
					disabled={
						isSubmitting || code.join("").length !== CODE_LENGTH
					}
					onClick={() => submitCode(code.join(""))}
				>
					{isSubmitting && <Spinner className="mr-2 size-4" />}
					Verificar
				</Button>
			</CardFooter>
		</Card>
	);
}
