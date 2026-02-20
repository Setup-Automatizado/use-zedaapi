"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
	CardDescription,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { CreditCard, QrCode, FileText } from "lucide-react";

type PaymentMethodType = "stripe" | "pix" | "boleto";

interface PaymentMethodFormProps {
	onSelect: (method: PaymentMethodType) => void;
	loading?: boolean;
	selectedMethod?: PaymentMethodType;
}

const methods: Array<{
	id: PaymentMethodType;
	label: string;
	description: string;
	icon: typeof CreditCard;
}> = [
	{
		id: "stripe",
		label: "Cartao de credito",
		description: "Pagamento recorrente via cartao. Aprovacao imediata.",
		icon: CreditCard,
	},
	{
		id: "pix",
		label: "PIX",
		description: "Pagamento instantaneo via PIX. QR Code gerado na hora.",
		icon: QrCode,
	},
	{
		id: "boleto",
		label: "Boleto bancario",
		description:
			"Boleto hibrido com QR Code PIX. Vencimento em 3 dias uteis.",
		icon: FileText,
	},
];

export function PaymentMethodForm({
	onSelect,
	loading = false,
	selectedMethod,
}: PaymentMethodFormProps) {
	const [selected, setSelected] = useState<PaymentMethodType>(
		selectedMethod || "stripe",
	);

	return (
		<div className="space-y-4">
			<div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
				{methods.map((method) => {
					const Icon = method.icon;
					const isSelected = selected === method.id;

					return (
						<button
							key={method.id}
							type="button"
							disabled={loading}
							onClick={() => setSelected(method.id)}
							className={cn(
								"flex flex-col items-start gap-2 rounded-xl border p-4 text-left transition-all duration-150 outline-none",
								"hover:border-primary/50 hover:bg-muted/50",
								"focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
								isSelected &&
									"border-primary bg-primary/5 ring-1 ring-primary",
								loading && "opacity-50 cursor-not-allowed",
							)}
						>
							<Icon
								className={cn(
									"size-5",
									isSelected
										? "text-primary"
										: "text-muted-foreground",
								)}
							/>
							<div>
								<p className="font-medium text-sm">
									{method.label}
								</p>
								<p className="text-xs text-muted-foreground mt-0.5">
									{method.description}
								</p>
							</div>
						</button>
					);
				})}
			</div>

			<Button
				className="w-full"
				disabled={loading}
				onClick={() => onSelect(selected)}
			>
				{loading ? "Processando..." : "Continuar com pagamento"}
			</Button>
		</div>
	);
}
