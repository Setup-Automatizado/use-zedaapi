"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/shared/empty-state";
import { ExternalLink, Copy, FileText } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

interface Invoice {
	id: string;
	amount: number | string;
	currency: string;
	status: string;
	paidAt: string | Date | null;
	dueDate: string | Date | null;
	paymentMethod: string | null;
	pdfUrl: string | null;
	createdAt: string | Date;
	subscription?: {
		plan: {
			name: string;
			slug: string;
		};
	} | null;
	sicrediCharge?: {
		type: string;
		status: string;
		pixCopiaECola: string | null;
		linhaDigitavel: string | null;
	} | null;
}

interface InvoiceTableProps {
	invoices: Invoice[];
	onRequestPayment?: (
		invoiceId: string,
		type: "pix" | "boleto_hibrido",
	) => void;
}

const statusConfig: Record<string, { label: string; className: string }> = {
	paid: { label: "Pago", className: "bg-primary/10 text-primary" },
	pending: { label: "Pendente", className: "bg-chart-2/10 text-chart-2" },
	draft: { label: "Rascunho", className: "bg-muted text-muted-foreground" },
	overdue: {
		label: "Vencida",
		className: "bg-destructive/10 text-destructive",
	},
	canceled: {
		label: "Cancelada",
		className: "bg-muted text-muted-foreground",
	},
	failed: {
		label: "Falhou",
		className: "bg-destructive/10 text-destructive",
	},
};

function formatDate(date: string | Date | null): string {
	if (!date) return "-";
	return new Intl.DateTimeFormat("pt-BR", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
	}).format(new Date(date));
}

function formatCurrency(amount: number | string, currency = "BRL"): string {
	const num = typeof amount === "string" ? parseFloat(amount) : amount;
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency,
	}).format(num);
}

async function copyToClipboard(text: string) {
	try {
		await navigator.clipboard.writeText(text);
		toast.success("Copiado para a area de transferencia");
	} catch {
		const textarea = document.createElement("textarea");
		textarea.value = text;
		document.body.appendChild(textarea);
		textarea.select();
		document.execCommand("copy");
		document.body.removeChild(textarea);
		toast.success("Copiado para a area de transferencia");
	}
}

export function InvoiceTable({
	invoices,
	onRequestPayment,
}: InvoiceTableProps) {
	if (invoices.length === 0) {
		return (
			<EmptyState
				icon={FileText}
				title="Nenhuma fatura encontrada"
				description="Suas faturas aparecerao aqui quando voce tiver uma assinatura ativa."
			/>
		);
	}

	return (
		<div className="overflow-hidden rounded-xl border">
			<Table>
				<TableHeader>
					<TableRow className="bg-muted/30 hover:bg-muted/30">
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Data
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Plano
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Valor
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Método
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Status
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide text-right">
							Acoes
						</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{invoices.map((invoice) => {
						const status = statusConfig[invoice.status] ?? {
							label: invoice.status,
							className: "text-muted-foreground",
						};

						return (
							<TableRow
								key={invoice.id}
								className="transition-colors duration-100"
							>
								<TableCell className="tabular-nums">
									{formatDate(invoice.createdAt)}
								</TableCell>
								<TableCell>
									{invoice.subscription?.plan.name ?? "-"}
								</TableCell>
								<TableCell className="font-medium tabular-nums">
									{formatCurrency(
										invoice.amount,
										invoice.currency,
									)}
								</TableCell>
								<TableCell>
									{invoice.paymentMethod === "stripe"
										? "Cartao"
										: invoice.paymentMethod === "pix"
											? "PIX"
											: invoice.paymentMethod === "boleto"
												? "Boleto"
												: (invoice.paymentMethod ??
													"-")}
								</TableCell>
								<TableCell>
									<Badge
										variant="secondary"
										className={cn(
											"font-medium",
											status.className,
										)}
									>
										{status.label}
									</Badge>
								</TableCell>
								<TableCell className="text-right">
									<div className="flex items-center justify-end gap-1">
										{invoice.pdfUrl && (
											<Button
												variant="ghost"
												size="icon-xs"
												asChild
											>
												<a
													href={invoice.pdfUrl}
													target="_blank"
													rel="noopener noreferrer"
												>
													<ExternalLink className="size-3.5" />
												</a>
											</Button>
										)}

										{invoice.sicrediCharge?.pixCopiaECola &&
											invoice.status !== "paid" && (
												<Button
													variant="ghost"
													size="icon-xs"
													onClick={() =>
														copyToClipboard(
															invoice
																.sicrediCharge!
																.pixCopiaECola!,
														)
													}
													title="Copiar PIX"
												>
													<Copy className="size-3.5" />
												</Button>
											)}

										{invoice.status === "pending" &&
											!invoice.sicrediCharge &&
											onRequestPayment && (
												<Button
													variant="outline"
													size="xs"
													onClick={() =>
														onRequestPayment(
															invoice.id,
															"pix",
														)
													}
												>
													Pagar com PIX
												</Button>
											)}
									</div>
								</TableCell>
							</TableRow>
						);
					})}
				</TableBody>
			</Table>
		</div>
	);
}
