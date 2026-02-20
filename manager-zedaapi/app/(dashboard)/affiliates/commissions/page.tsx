import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import {
	getAffiliate,
	getCommissions,
} from "@/server/services/affiliate-service";
import { CommissionTable } from "@/components/affiliates/commission-table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { redirect } from "next/navigation";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Comissoes | Zé da API Manager",
};

interface Props {
	searchParams: Promise<{ page?: string }>;
}

export default async function CommissionsPage({ searchParams }: Props) {
	const session = await requireAuth();
	const affiliate = await getAffiliate(session.user.id);

	if (!affiliate) {
		redirect("/affiliates");
	}

	const params = await searchParams;
	const page = Number(params.page) || 1;
	const result = await getCommissions(affiliate.id, page);

	const totalPages = Math.ceil(result.total / result.pageSize);

	return (
		<div className="mx-auto max-w-4xl space-y-6 p-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-semibold">Comissões</h1>
					<p className="text-muted-foreground mt-1">
						Histórico de comissões geradas pelas suas indicações.
					</p>
				</div>
				<Link
					href="/affiliates"
					className="text-sm text-muted-foreground hover:text-foreground"
				>
					Voltar
				</Link>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-medium">
						{result.total} comissão{result.total !== 1 ? "ões" : ""}{" "}
						encontrada{result.total !== 1 ? "s" : ""}
					</CardTitle>
				</CardHeader>
				<CardContent>
					<CommissionTable commissions={result.items as never[]} />

					{totalPages > 1 && (
						<div className="flex justify-center gap-2 mt-4">
							{page > 1 && (
								<Link
									href={`/affiliates/commissions?page=${page - 1}`}
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
									href={`/affiliates/commissions?page=${page + 1}`}
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
