import { Suspense } from "react";
import type { Metadata } from "next";
import { requireAdmin } from "@/lib/auth-server";
import { db } from "@/lib/db";
import { getActiveConfig } from "@/lib/services/nfse/service";
import { NfeDashboard } from "./nfe-dashboard";
import {
	CardsSkeleton,
	TableSkeleton,
} from "@/components/shared/loading-skeleton";
import { PageHeader } from "@/components/shared/page-header";

export const metadata: Metadata = {
	title: "NFS-e Nacional | Admin Zé da API Manager",
};

export default async function NfePage() {
	await requireAdmin();

	return (
		<div className="space-y-6">
			<PageHeader
				title="NFS-e Nacional"
				description="Gerenciamento de notas fiscais de serviço."
			/>

			<Suspense
				fallback={
					<>
						<CardsSkeleton count={5} />
						<TableSkeleton />
					</>
				}
			>
				<NfeContent />
			</Suspense>
		</div>
	);
}

async function NfeContent() {
	const config = await getActiveConfig();

	const recentInvoices = await db.invoice.findMany({
		where: { nfseStatus: { not: null } },
		orderBy: { updatedAt: "desc" },
		take: 50,
		select: {
			id: true,
			amount: true,
			currency: true,
			status: true,
			nfseStatus: true,
			nfseNumber: true,
			nfseProtocol: true,
			nfseXmlUrl: true,
			nfsePdfUrl: true,
			nfseIssuedAt: true,
			nfseError: true,
			nfseCanceledAt: true,
			createdAt: true,
			updatedAt: true,
			user: {
				select: {
					name: true,
					email: true,
					cpfCnpj: true,
				},
			},
		},
	});

	const stats = await db.invoice.groupBy({
		by: ["nfseStatus"],
		_count: true,
		where: { nfseStatus: { not: null } },
	});

	const statsMap: Record<string, number> = {};
	for (const s of stats) {
		if (s.nfseStatus) {
			statsMap[s.nfseStatus] = s._count;
		}
	}

	const safeConfig = config
		? {
				id: config.id,
				active: config.active,
				cnpj: config.cnpj,
				inscricaoMunicipal: config.inscricaoMunicipal,
				codigoMunicipio: config.codigoMunicipio,
				uf: config.uf,
				certificateExpiresAt: config.certificateExpiresAt.toISOString(),
				codigoServico: config.codigoServico,
				cnae: config.cnae,
				aliquotaIss: config.aliquotaIss,
				descricaoServico: config.descricaoServico,
				codigoServicoPf: config.codigoServicoPf,
				cnaePf: config.cnaePf,
				aliquotaIssPf: config.aliquotaIssPf,
				descricaoServicoPf: config.descricaoServicoPf,
				ambiente: config.ambiente,
			}
		: null;

	const serializedInvoices = recentInvoices.map((inv) => ({
		...inv,
		amount: Number(inv.amount),
		createdAt: inv.createdAt.toISOString(),
		updatedAt: inv.updatedAt.toISOString(),
		nfseIssuedAt: inv.nfseIssuedAt?.toISOString() || null,
		nfseCanceledAt: inv.nfseCanceledAt?.toISOString() || null,
	}));

	return (
		<NfeDashboard
			config={safeConfig}
			invoices={serializedInvoices}
			stats={statsMap}
		/>
	);
}
