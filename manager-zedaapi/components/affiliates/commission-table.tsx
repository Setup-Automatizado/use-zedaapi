"use client";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { EmptyState } from "@/components/shared/empty-state";
import { Coins } from "lucide-react";

interface Commission {
	id: string;
	amount: string | number;
	status: string;
	createdAt: string;
	paidAt: string | null;
	referral: {
		referredUser: {
			name: string;
			email: string;
		};
	};
	invoice: {
		id: string;
		amount: string | number;
		paidAt: string | null;
	} | null;
}

interface CommissionTableProps {
	commissions: Commission[];
}

const statusConfig: Record<string, { label: string; className: string }> = {
	pending: {
		label: "Pendente",
		className: "bg-chart-2/10 text-chart-2",
	},
	approved: {
		label: "Aprovada",
		className: "bg-primary/10 text-primary",
	},
	paid: {
		label: "Paga",
		className: "bg-muted text-muted-foreground",
	},
};

function formatCurrency(value: unknown): string {
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(Number(value));
}

function formatDate(date: string | Date): string {
	return new Date(date).toLocaleDateString("pt-BR");
}

export function CommissionTable({ commissions }: CommissionTableProps) {
	if (commissions.length === 0) {
		return (
			<EmptyState
				icon={Coins}
				title="Nenhuma comissao encontrada"
				description="Comissoes aparecem quando seus indicados efetuam pagamentos."
			/>
		);
	}

	return (
		<div className="overflow-hidden rounded-xl border">
			<Table>
				<TableHeader>
					<TableRow className="bg-muted/30 hover:bg-muted/30">
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Indicado
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Valor Fatura
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Comissao
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Status
						</TableHead>
						<TableHead className="text-xs font-medium uppercase tracking-wide">
							Data
						</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{commissions.map((c) => (
						<TableRow
							key={c.id}
							className="transition-colors duration-100"
						>
							<TableCell>
								<div>
									<p className="text-sm font-medium">
										{c.referral.referredUser.name}
									</p>
									<p className="text-xs text-muted-foreground">
										{c.referral.referredUser.email}
									</p>
								</div>
							</TableCell>
							<TableCell className="tabular-nums">
								{c.invoice
									? formatCurrency(c.invoice.amount)
									: "\u2014"}
							</TableCell>
							<TableCell className="font-medium tabular-nums">
								{formatCurrency(c.amount)}
							</TableCell>
							<TableCell>
								<Badge
									variant="secondary"
									className={cn(
										"font-medium",
										statusConfig[c.status]?.className ??
											"bg-muted text-muted-foreground",
									)}
								>
									{statusConfig[c.status]?.label ?? c.status}
								</Badge>
							</TableCell>
							<TableCell className="text-muted-foreground tabular-nums">
								{formatDate(c.createdAt)}
							</TableCell>
						</TableRow>
					))}
				</TableBody>
			</Table>
		</div>
	);
}
