"use client";

import { useMemo, useState, useTransition } from "react";
import { useSearchParams } from "next/navigation";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion, AnimatePresence } from "framer-motion";
import {
	CheckCircle2Icon,
	SendIcon,
	AlertCircleIcon,
	RotateCcwIcon,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Spinner } from "@/components/ui/spinner";
import { submitContactForm } from "@/server/actions/contact";

// =============================================================================
// Schema
// =============================================================================

const contactFormSchema = z.object({
	nome: z.string().min(2, "Nome deve ter pelo menos 2 caracteres"),
	email: z.string().email("E-mail inválido"),
	whatsapp: z
		.string()
		.regex(/^\(?\d{2}\)?\s?\d{4,5}-?\d{4}$/, "Formato inválido")
		.or(z.literal(""))
		.optional(),
	empresa: z.string().optional(),
	assunto: z.enum([
		"comercial",
		"suporte",
		"parceria",
		"financeiro",
		"outro",
	]),
	mensagem: z.string().min(10, "Mensagem deve ter pelo menos 10 caracteres"),
	preferencia_contato: z.enum(["email", "whatsapp", "ambos"]),
});

type ContactFormValues = z.infer<typeof contactFormSchema>;

const ASSUNTO_OPTIONS = [
	{ value: "comercial", label: "Comercial" },
	{ value: "suporte", label: "Suporte Técnico" },
	{ value: "parceria", label: "Parceria" },
	{ value: "financeiro", label: "Financeiro" },
	{ value: "outro", label: "Outro" },
] as const;

// =============================================================================
// Phone mask utility
// =============================================================================

function applyPhoneMask(value: string): string {
	const digits = value.replace(/\D/g, "").slice(0, 11);
	if (digits.length <= 2) return digits;
	if (digits.length <= 7) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
	return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7)}`;
}

// =============================================================================
// Component
// =============================================================================

export function ContactForm() {
	const searchParams = useSearchParams();
	const [isPending, startTransition] = useTransition();
	const [formState, setFormState] = useState<"idle" | "success" | "error">(
		"idle",
	);
	const [serverError, setServerError] = useState<string | null>(null);

	const utmParams = useMemo(
		() => ({
			utm_source: searchParams.get("utm_source") ?? "",
			utm_medium: searchParams.get("utm_medium") ?? "",
			utm_campaign: searchParams.get("utm_campaign") ?? "",
			utm_term: searchParams.get("utm_term") ?? "",
			utm_content: searchParams.get("utm_content") ?? "",
		}),
		[searchParams],
	);

	const {
		register,
		handleSubmit,
		setValue,
		control,
		reset,
		formState: { errors },
	} = useForm<ContactFormValues>({
		resolver: zodResolver(contactFormSchema),
		defaultValues: {
			nome: "",
			email: "",
			whatsapp: "",
			empresa: "",
			assunto: "comercial",
			mensagem: "",
			preferencia_contato: "email",
		},
	});

	const assuntoValue = useWatch({ control, name: "assunto" });
	const preferenciaValue = useWatch({ control, name: "preferencia_contato" });

	function onSubmit(data: ContactFormValues) {
		setServerError(null);

		startTransition(async () => {
			const fd = new FormData();
			fd.set("nome", data.nome);
			fd.set("email", data.email);
			fd.set("whatsapp", data.whatsapp ?? "");
			fd.set("empresa", data.empresa ?? "");
			fd.set("assunto", data.assunto);
			fd.set("mensagem", data.mensagem);
			fd.set("preferencia_contato", data.preferencia_contato);
			fd.set("utm_source", utmParams.utm_source);
			fd.set("utm_medium", utmParams.utm_medium);
			fd.set("utm_campaign", utmParams.utm_campaign);
			fd.set("utm_term", utmParams.utm_term);
			fd.set("utm_content", utmParams.utm_content);
			fd.set("page_url", window.location.href);
			fd.set("referrer", document.referrer);

			const result = await submitContactForm(fd);

			if (result.success) {
				setFormState("success");
			} else {
				setServerError(
					result.error ?? "Erro ao enviar mensagem. Tente novamente.",
				);
				setFormState("error");
			}
		});
	}

	function handleRetry() {
		setFormState("idle");
		setServerError(null);
	}

	function handleNewMessage() {
		setFormState("idle");
		setServerError(null);
		reset();
	}

	return (
		<div className="relative">
			<AnimatePresence mode="wait">
				{formState === "success" ? (
					<motion.div
						key="success"
						initial={{ opacity: 0, scale: 0.95 }}
						animate={{ opacity: 1, scale: 1 }}
						exit={{ opacity: 0, scale: 0.95 }}
						transition={{ duration: 0.3, ease: "easeOut" }}
						className="flex flex-col items-center justify-center gap-5 rounded-2xl border border-border bg-card p-10 text-center shadow-sm"
					>
						<motion.div
							initial={{ scale: 0 }}
							animate={{ scale: 1 }}
							transition={{
								type: "spring",
								stiffness: 300,
								damping: 20,
								delay: 0.1,
							}}
							className="flex size-16 items-center justify-center rounded-full bg-green-500/10"
						>
							<CheckCircle2Icon className="size-8 text-green-500" />
						</motion.div>
						<div className="space-y-2">
							<h3 className="text-xl font-semibold text-foreground">
								Mensagem enviada!
							</h3>
							<p className="max-w-sm text-sm leading-relaxed text-muted-foreground">
								Obrigado pelo contato. Nossa equipe retornará em
								breve pelo canal de sua preferência.
							</p>
						</div>
						<Button
							variant="outline"
							size="sm"
							onClick={handleNewMessage}
							className="mt-2"
						>
							Enviar outra mensagem
						</Button>
					</motion.div>
				) : (
					<motion.form
						key="form"
						initial={{ opacity: 0, y: 8 }}
						animate={{ opacity: 1, y: 0 }}
						exit={{ opacity: 0, y: -8 }}
						transition={{ duration: 0.3 }}
						onSubmit={handleSubmit(onSubmit)}
						className="flex flex-col gap-5 rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8"
					>
						<div className="space-y-1">
							<h3 className="text-lg font-semibold text-foreground">
								Envie sua mensagem
							</h3>
							<p className="text-sm text-muted-foreground">
								Preencha o formulário abaixo e retornaremos em
								breve.
							</p>
						</div>

						{serverError && (
							<motion.div
								initial={{ opacity: 0, height: 0 }}
								animate={{ opacity: 1, height: "auto" }}
								className="flex items-start gap-2.5 rounded-xl bg-destructive/10 px-4 py-3"
							>
								<AlertCircleIcon className="mt-0.5 size-4 shrink-0 text-destructive" />
								<div className="flex-1">
									<p className="text-sm text-destructive">
										{serverError}
									</p>
									<button
										type="button"
										onClick={handleRetry}
										className="mt-1 inline-flex items-center gap-1 text-xs font-medium text-destructive underline-offset-4 hover:underline"
									>
										<RotateCcwIcon className="size-3" />
										Tentar novamente
									</button>
								</div>
							</motion.div>
						)}

						{/* Nome */}
						<div className="flex flex-col gap-1.5">
							<Label htmlFor="nome">
								Nome <span className="text-destructive">*</span>
							</Label>
							<Input
								id="nome"
								placeholder="Seu nome completo"
								aria-invalid={!!errors.nome}
								className={cn(
									errors.nome &&
										"border-destructive focus-visible:ring-destructive",
								)}
								{...register("nome")}
							/>
							{errors.nome && (
								<p className="text-xs text-destructive">
									{errors.nome.message}
								</p>
							)}
						</div>

						{/* Email */}
						<div className="flex flex-col gap-1.5">
							<Label htmlFor="email">
								E-mail{" "}
								<span className="text-destructive">*</span>
							</Label>
							<Input
								id="email"
								type="email"
								placeholder="seu@email.com"
								aria-invalid={!!errors.email}
								className={cn(
									errors.email &&
										"border-destructive focus-visible:ring-destructive",
								)}
								{...register("email")}
							/>
							{errors.email && (
								<p className="text-xs text-destructive">
									{errors.email.message}
								</p>
							)}
						</div>

						{/* WhatsApp + Empresa */}
						<div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
							<div className="flex flex-col gap-1.5">
								<Label htmlFor="whatsapp">WhatsApp</Label>
								<Input
									id="whatsapp"
									placeholder="(41) 99999-9999"
									aria-invalid={!!errors.whatsapp}
									className={cn(
										errors.whatsapp &&
											"border-destructive focus-visible:ring-destructive",
									)}
									{...register("whatsapp", {
										onChange: (e) => {
											e.target.value = applyPhoneMask(
												e.target.value,
											);
										},
									})}
								/>
								{errors.whatsapp && (
									<p className="text-xs text-destructive">
										{errors.whatsapp.message}
									</p>
								)}
							</div>
							<div className="flex flex-col gap-1.5">
								<Label htmlFor="empresa">Empresa</Label>
								<Input
									id="empresa"
									placeholder="Nome da empresa"
									{...register("empresa")}
								/>
							</div>
						</div>

						{/* Assunto */}
						<div className="flex flex-col gap-1.5">
							<Label htmlFor="assunto">
								Assunto{" "}
								<span className="text-destructive">*</span>
							</Label>
							<Select
								value={assuntoValue}
								onValueChange={(v) =>
									setValue(
										"assunto",
										v as ContactFormValues["assunto"],
										{ shouldValidate: true },
									)
								}
							>
								<SelectTrigger
									className={cn(
										"w-full",
										errors.assunto &&
											"border-destructive focus-visible:ring-destructive",
									)}
								>
									<SelectValue placeholder="Selecione o assunto" />
								</SelectTrigger>
								<SelectContent>
									{ASSUNTO_OPTIONS.map((opt) => (
										<SelectItem
											key={opt.value}
											value={opt.value}
										>
											{opt.label}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
							{errors.assunto && (
								<p className="text-xs text-destructive">
									{errors.assunto.message}
								</p>
							)}
						</div>

						{/* Mensagem */}
						<div className="flex flex-col gap-1.5">
							<Label htmlFor="mensagem">
								Mensagem{" "}
								<span className="text-destructive">*</span>
							</Label>
							<Textarea
								id="mensagem"
								placeholder="Descreva como podemos ajudar..."
								rows={5}
								aria-invalid={!!errors.mensagem}
								className={cn(
									"resize-none",
									errors.mensagem &&
										"border-destructive focus-visible:ring-destructive",
								)}
								{...register("mensagem")}
							/>
							{errors.mensagem && (
								<p className="text-xs text-destructive">
									{errors.mensagem.message}
								</p>
							)}
						</div>

						{/* Preferencia de contato */}
						<div className="flex flex-col gap-2.5">
							<Label>Prefere ser contatado por:</Label>
							<RadioGroup
								value={preferenciaValue}
								onValueChange={(v) =>
									setValue(
										"preferencia_contato",
										v as ContactFormValues["preferencia_contato"],
										{ shouldValidate: true },
									)
								}
								className="flex flex-wrap gap-3"
							>
								{[
									{
										value: "email",
										label: "E-mail",
									},
									{
										value: "whatsapp",
										label: "WhatsApp",
									},
									{
										value: "ambos",
										label: "Ambos",
									},
								].map((option) => (
									<label
										key={option.value}
										htmlFor={`pref-${option.value}`}
										className={cn(
											"flex cursor-pointer items-center gap-2 rounded-xl border px-4 py-2.5 text-sm transition-all duration-150",
											preferenciaValue === option.value
												? "border-primary bg-primary/5 text-foreground"
												: "border-border bg-background text-muted-foreground hover:border-primary/50 hover:bg-muted/50",
										)}
									>
										<RadioGroupItem
											value={option.value}
											id={`pref-${option.value}`}
										/>
										{option.label}
									</label>
								))}
							</RadioGroup>
						</div>

						{/* Submit */}
						<Button
							type="submit"
							size="lg"
							disabled={isPending}
							className="mt-2 gap-2 rounded-4xl"
						>
							{isPending ? (
								<>
									<Spinner className="size-4" />
									Enviando...
								</>
							) : (
								<>
									<SendIcon className="size-4" />
									Enviar mensagem
								</>
							)}
						</Button>
					</motion.form>
				)}
			</AnimatePresence>
		</div>
	);
}
