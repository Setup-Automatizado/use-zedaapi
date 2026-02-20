"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { CheckCircle2Icon, LoaderIcon, AlertTriangleIcon } from "lucide-react";

const reasons = [
	{ value: "not-using", label: "Não utilizo mais o serviço" },
	{ value: "privacy-concern", label: "Preocupação com privacidade" },
	{ value: "switching-provider", label: "Mudando para outro provedor" },
	{ value: "cost", label: "Custo do serviço" },
	{ value: "other", label: "Outro motivo" },
] as const;

export function DataDeletionForm() {
	const [status, setStatus] = useState<
		"idle" | "submitting" | "success" | "error"
	>("idle");
	const [errorMessage, setErrorMessage] = useState("");
	const [confirmed, setConfirmed] = useState(false);

	if (status === "success") {
		return (
			<div className="my-8 rounded-xl border border-primary/20 bg-primary/5 p-6">
				<div className="flex items-start gap-3">
					<CheckCircle2Icon className="mt-0.5 size-5 shrink-0 text-primary" />
					<div>
						<h3 className="text-base font-semibold text-foreground">
							Solicitação enviada com sucesso
						</h3>
						<p className="mt-1 text-sm text-muted-foreground">
							Recebemos sua solicitação de exclusão de dados. Você
							receberá um e-mail de confirmação em breve. O
							processo de exclusão será concluído em até{" "}
							<strong>30 dias corridos</strong>.
						</p>
						<p className="mt-2 text-sm text-muted-foreground">
							Caso tenha dúvidas, entre em contato com nosso DPO
							em{" "}
							<a
								href="mailto:privacidade@zedaapi.com"
								className="text-primary underline underline-offset-4"
							>
								privacidade@zedaapi.com
							</a>
							.
						</p>
					</div>
				</div>
			</div>
		);
	}

	async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
		e.preventDefault();
		setStatus("submitting");
		setErrorMessage("");

		const formData = new FormData(e.currentTarget);
		const data = {
			name: formData.get("name") as string,
			email: formData.get("email") as string,
			document: formData.get("document") as string,
			reason: formData.get("reason") as string,
			details: formData.get("details") as string,
		};

		try {
			const response = await fetch("/api/data-deletion", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(data),
			});

			if (!response.ok) {
				const body = await response
					.json()
					.catch(() => ({
						message: "Erro ao processar solicitação.",
					}));
				throw new Error(
					body.message || "Erro ao processar solicitação.",
				);
			}

			setStatus("success");
		} catch (err) {
			setStatus("error");
			setErrorMessage(
				err instanceof Error
					? err.message
					: "Ocorreu um erro inesperado. Tente novamente.",
			);
		}
	}

	return (
		<form
			onSubmit={handleSubmit}
			className="my-8 space-y-6 rounded-xl border border-border bg-card p-6"
		>
			{/* Name */}
			<div className="space-y-2">
				<Label htmlFor="deletion-name">
					Nome completo <span className="text-destructive">*</span>
				</Label>
				<Input
					id="deletion-name"
					name="name"
					type="text"
					required
					placeholder="Seu nome completo"
					autoComplete="name"
				/>
			</div>

			{/* Email */}
			<div className="space-y-2">
				<Label htmlFor="deletion-email">
					E-mail cadastrado{" "}
					<span className="text-destructive">*</span>
				</Label>
				<Input
					id="deletion-email"
					name="email"
					type="email"
					required
					placeholder="seu@email.com"
					autoComplete="email"
				/>
				<p className="text-xs text-muted-foreground">
					Informe o e-mail utilizado no cadastro da Plataforma.
				</p>
			</div>

			{/* CPF/CNPJ */}
			<div className="space-y-2">
				<Label htmlFor="deletion-document">
					CPF ou CNPJ (opcional)
				</Label>
				<Input
					id="deletion-document"
					name="document"
					type="text"
					placeholder="000.000.000-00 ou 00.000.000/0000-00"
				/>
				<p className="text-xs text-muted-foreground">
					Fornecer o CPF/CNPJ ajuda a agilizar a verificação de
					identidade.
				</p>
			</div>

			{/* Reason */}
			<div className="space-y-2">
				<Label htmlFor="deletion-reason">
					Motivo da solicitação{" "}
					<span className="text-destructive">*</span>
				</Label>
				<Select name="reason" required>
					<SelectTrigger className="w-full" id="deletion-reason">
						<SelectValue placeholder="Selecione um motivo" />
					</SelectTrigger>
					<SelectContent>
						{reasons.map((reason) => (
							<SelectItem key={reason.value} value={reason.value}>
								{reason.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>

			{/* Details */}
			<div className="space-y-2">
				<Label htmlFor="deletion-details">
					Detalhes adicionais (opcional)
				</Label>
				<textarea
					id="deletion-details"
					name="details"
					rows={4}
					placeholder="Descreva informações adicionais que possam nos ajudar a processar sua solicitação..."
					className="bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 w-full rounded-xl border px-3 py-2 text-sm transition-colors focus-visible:ring-[3px] placeholder:text-muted-foreground outline-none resize-none"
				/>
			</div>

			{/* Confirmation checkbox */}
			<div className="flex items-start gap-3 rounded-lg border border-destructive/20 bg-destructive/5 p-4">
				<Checkbox
					id="deletion-confirm"
					checked={confirmed}
					onCheckedChange={(checked) =>
						setConfirmed(checked === true)
					}
					className="mt-0.5"
				/>
				<Label
					htmlFor="deletion-confirm"
					className="text-sm font-normal leading-relaxed text-muted-foreground"
				>
					Entendo que esta ação é{" "}
					<strong className="text-foreground">
						permanente e irreversível
					</strong>
					. Todos os meus dados, instâncias WhatsApp, configurações e
					integrações serão permanentemente excluídos. Registros
					financeiros e fiscais serão retidos conforme exigido por
					lei.
				</Label>
			</div>

			{/* Error message */}
			{status === "error" && errorMessage && (
				<div className="flex items-start gap-3 rounded-lg border border-destructive/20 bg-destructive/5 p-4">
					<AlertTriangleIcon className="mt-0.5 size-4 shrink-0 text-destructive" />
					<p className="text-sm text-destructive">{errorMessage}</p>
				</div>
			)}

			{/* Submit */}
			<Button
				type="submit"
				variant="destructive"
				disabled={!confirmed || status === "submitting"}
				className="w-full sm:w-auto"
			>
				{status === "submitting" ? (
					<>
						<LoaderIcon className="size-4 animate-spin" />
						Enviando solicitação...
					</>
				) : (
					"Solicitar exclusão de dados"
				)}
			</Button>
		</form>
	);
}
