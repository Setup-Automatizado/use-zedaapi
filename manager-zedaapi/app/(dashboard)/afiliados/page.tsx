import type { Metadata } from "next";
import { requireAuth } from "@/lib/auth-server";
import {
	getAffiliate,
	getAffiliateStats,
} from "@/server/services/affiliate-service";
import { registerAffiliate } from "@/server/services/affiliate-service";
import { ReferralLink } from "@/components/affiliates/referral-link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/shared/page-header";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Afiliados | Zé da API Manager",
};

export default async function AffiliatesPage() {
	const session = await requireAuth();
	let affiliate = await getAffiliate(session.user.id);

	// Auto-register if user navigates here and isn't an affiliate yet
	if (!affiliate) {
		try {
			await registerAffiliate(session.user.id);
			affiliate = await getAffiliate(session.user.id);
		} catch {
			return (
				<div className="mx-auto max-w-4xl space-y-6 p-6">
					<h1 className="text-2xl font-semibold">
						Programa de Afiliados
					</h1>
					<p className="text-muted-foreground">
						Não foi possível registrar como afiliado. Tente
						novamente mais tarde.
					</p>
				</div>
			);
		}
	}

	if (!affiliate) return null;

	const stats = await getAffiliateStats(affiliate.id);

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat("pt-BR", {
			style: "currency",
			currency: "BRL",
		}).format(value);
	}

	return (
		<div className="mx-auto max-w-4xl space-y-6 p-6">
			<PageHeader
				title="Programa de Afiliados"
				description={`Indique novos usuários e receba ${(Number(affiliate.commissionRate) * 100).toFixed(0)}% de comissão sobre cada pagamento.`}
			/>

			<ReferralLink code={affiliate.code} />

			<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Total Ganho
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{formatCurrency(
								stats.paidAmount +
									stats.approvedAmount +
									stats.pendingAmount,
							)}
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Pendente
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{formatCurrency(stats.pendingAmount)}
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Indicações Convertidas
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{stats.convertedReferrals}/{stats.totalReferrals}
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Taxa de Conversão
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{stats.conversionRate.toFixed(1)}%
						</p>
					</CardContent>
				</Card>
			</div>

			<div className="grid gap-4 sm:grid-cols-2">
				<Link href="/afiliados/comissoes">
					<Card className="transition-colors hover:bg-muted/50 cursor-pointer">
						<CardHeader>
							<CardTitle className="text-sm">Comissões</CardTitle>
						</CardHeader>
						<CardContent>
							<p className="text-muted-foreground text-sm">
								Veja o histórico de todas as comissões geradas
								pelas suas indicações.
							</p>
						</CardContent>
					</Card>
				</Link>
				<Link href="/afiliados/pagamentos">
					<Card className="transition-colors hover:bg-muted/50 cursor-pointer">
						<CardHeader>
							<CardTitle className="text-sm">
								Pagamentos
							</CardTitle>
						</CardHeader>
						<CardContent>
							<p className="text-muted-foreground text-sm">
								Acompanhe seus saques e solicitações de
								pagamento.
							</p>
						</CardContent>
					</Card>
				</Link>
			</div>
		</div>
	);
}
