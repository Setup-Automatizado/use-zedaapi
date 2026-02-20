import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import { getAffiliate, getPayouts } from "@/server/services/affiliate-service";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { PageHeader } from "@/components/shared/page-header";
import { redirect } from "next/navigation";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Pagamentos | Zé da API Manager",
};

interface Props {
	searchParams: Promise<{ page?: string }>;
}

const statusVariants: Record<
	string,
	"default" | "secondary" | "destructive" | "outline"
> = {
	pending: "secondary",
	processing: "default",
	completed: "outline",
	failed: "destructive",
};

const statusLabels: Record<string, string> = {
	pending: "Pendente",
	processing: "Processando",
	completed: "Concluído",
	failed: "Falhou",
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

export default async function PayoutsPage({ searchParams }: Props) {
	const session = await requireAuth();
	const affiliate = await getAffiliate(session.user.id);

	if (!affiliate) {
		redirect("/affiliates");
	}

	const params = await searchParams;
	const page = Number(params.page) || 1;
	const result = await getPayouts(affiliate.id, page);

	const totalPages = Math.ceil(result.total / result.pageSize);

	return (
		<div className="mx-auto max-w-4xl space-y-6 p-6">
			<PageHeader
				title="Pagamentos"
				description="Histórico de saques e pagamentos do programa de afiliados."
				backHref="/affiliates"
			/>

			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-medium">
						{result.total} pagamento{result.total !== 1 ? "s" : ""}{" "}
						encontrado{result.total !== 1 ? "s" : ""}
					</CardTitle>
				</CardHeader>
				<CardContent>
					{result.items.length === 0 ? (
						<div className="flex flex-col items-center justify-center py-12 text-center">
							<p className="text-muted-foreground">
								Nenhum pagamento encontrado.
							</p>
							<p className="text-xs text-muted-foreground mt-1">
								Solicite um saque quando tiver comissões
								aprovadas.
							</p>
						</div>
					) : (
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Valor</TableHead>
									<TableHead>Método</TableHead>
									<TableHead>Status</TableHead>
									<TableHead>Data</TableHead>
									<TableHead>Processado em</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{result.items.map((p) => (
									<TableRow key={p.id}>
										<TableCell className="font-medium">
											{formatCurrency(p.amount)}
										</TableCell>
										<TableCell className="uppercase text-xs">
											{p.method}
										</TableCell>
										<TableCell>
											<Badge
												variant={
													statusVariants[p.status] ||
													"secondary"
												}
											>
												{statusLabels[p.status] ||
													p.status}
											</Badge>
										</TableCell>
										<TableCell className="text-muted-foreground">
											{formatDate(p.createdAt)}
										</TableCell>
										<TableCell className="text-muted-foreground">
											{p.processedAt
												? formatDate(p.processedAt)
												: "—"}
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					)}

					{totalPages > 1 && (
						<div className="flex justify-center gap-2 mt-4">
							{page > 1 && (
								<Link
									href={`/affiliates/payouts?page=${page - 1}`}
									className="text-sm px-3 py-1 rounded border hover:bg-muted"
								>
									Anterior
								</Link>
							)}
							<span className="text-sm px-3 py-1 text-muted-foreground">
								Página {page} de {totalPages}
							</span>
							{page < totalPages && (
								<Link
									href={`/affiliates/payouts?page=${page + 1}`}
									className="text-sm px-3 py-1 rounded border hover:bg-muted"
								>
									Próxima
								</Link>
							)}
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
