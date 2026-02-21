import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { db } from "@/lib/db";
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
import {
	CardsSkeleton,
	TableSkeleton,
} from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";

export const metadata: Metadata = {
	title: "Afiliados | Admin Zé da API Manager",
};

function formatCurrency(value: unknown): string {
	return new Intl.NumberFormat("pt-BR", {
		style: "currency",
		currency: "BRL",
	}).format(Number(value));
}

function formatDate(date: Date): string {
	return date.toLocaleDateString("pt-BR");
}

const statusVariants: Record<
	string,
	"default" | "secondary" | "destructive" | "outline"
> = {
	active: "default",
	pending: "secondary",
	suspended: "destructive",
};

const statusLabels: Record<string, string> = {
	active: "Ativo",
	pending: "Pendente",
	suspended: "Suspenso",
};

export default async function AdminAffiliatesPage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="Gestão de Afiliados"
				description="Gerencie afiliados, comissões e pagamentos."
			/>

			<Suspense
				fallback={
					<>
						<CardsSkeleton count={3} />
						<TableSkeleton />
					</>
				}
			>
				<AffiliatesContent />
			</Suspense>
		</div>
	);
}

async function AffiliatesContent() {
	const affiliatesQuery = db.affiliate.findMany({
		include: {
			user: { select: { name: true, email: true } },
			_count: {
				select: {
					referrals: true,
					commissions: true,
					payouts: true,
				},
			},
		},
		orderBy: { createdAt: "desc" },
	});

	const pendingPayoutsQuery = db.payout.findMany({
		where: { status: "pending" },
		include: {
			affiliate: {
				include: { user: { select: { name: true, email: true } } },
			},
		},
		orderBy: { createdAt: "desc" },
	});

	const totalCommissionsQuery = db.commission.aggregate({
		where: { status: "pending" },
		_sum: { amount: true },
	});

	const [affiliates, pendingPayouts, totalCommissions] = await Promise.all([
		affiliatesQuery,
		pendingPayoutsQuery,
		totalCommissionsQuery,
	]);

	return (
		<>
			<div className="grid gap-4 sm:grid-cols-3">
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Total Afiliados
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{affiliates.length}
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Pagamentos Pendentes
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{pendingPayouts.length}
						</p>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-xs font-medium text-muted-foreground">
							Comissões Pendentes
						</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-bold">
							{formatCurrency(
								Number(totalCommissions._sum.amount || 0),
							)}
						</p>
					</CardContent>
				</Card>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-sm font-medium">
						Todos os Afiliados
					</CardTitle>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Afiliado</TableHead>
								<TableHead>Código</TableHead>
								<TableHead>Taxa</TableHead>
								<TableHead>Indicações</TableHead>
								<TableHead>Total Ganho</TableHead>
								<TableHead>Total Pago</TableHead>
								<TableHead>Status</TableHead>
								<TableHead>Desde</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{affiliates.map((a) => (
								<TableRow key={a.id}>
									<TableCell>
										<div>
											<p className="font-medium text-sm">
												{a.user.name}
											</p>
											<p className="text-xs text-muted-foreground">
												{a.user.email}
											</p>
										</div>
									</TableCell>
									<TableCell className="font-mono text-xs">
										{a.code}
									</TableCell>
									<TableCell>
										{(
											Number(a.commissionRate) * 100
										).toFixed(0)}
										%
									</TableCell>
									<TableCell>{a._count.referrals}</TableCell>
									<TableCell>
										{formatCurrency(a.totalEarnings)}
									</TableCell>
									<TableCell>
										{formatCurrency(a.totalPaid)}
									</TableCell>
									<TableCell>
										<Badge
											variant={
												statusVariants[a.status] ||
												"secondary"
											}
										>
											{statusLabels[a.status] || a.status}
										</Badge>
									</TableCell>
									<TableCell className="text-muted-foreground">
										{formatDate(a.createdAt)}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</CardContent>
			</Card>

			{pendingPayouts.length > 0 && (
				<Card>
					<CardHeader>
						<CardTitle className="text-sm font-medium">
							Pagamentos Pendentes de Aprovação
						</CardTitle>
					</CardHeader>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Afiliado</TableHead>
									<TableHead>Valor</TableHead>
									<TableHead>Método</TableHead>
									<TableHead>Data</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{pendingPayouts.map((p) => (
									<TableRow key={p.id}>
										<TableCell>
											{p.affiliate.user.name}
										</TableCell>
										<TableCell className="font-medium">
											{formatCurrency(p.amount)}
										</TableCell>
										<TableCell className="uppercase text-xs">
											{p.method}
										</TableCell>
										<TableCell className="text-muted-foreground">
											{formatDate(p.createdAt)}
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</CardContent>
				</Card>
			)}
		</>
	);
}
