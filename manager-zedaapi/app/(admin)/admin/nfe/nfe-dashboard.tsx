"use client";

import { useState, useTransition } from "react";
import {
	issueNfse,
	cancelNfseAction,
	retryNfse,
	getNfseStatus,
} from "@/server/actions/nfe";
import { Button } from "@/components/ui/button";

interface NfeConfig {
	id: string;
	active: boolean;
	cnpj: string;
	inscricaoMunicipal: string;
	codigoMunicipio: string;
	uf: string;
	certificateExpiresAt: string;
	codigoServico: string;
	cnae: string;
	aliquotaIss: number;
	descricaoServico: string;
	codigoServicoPf: string;
	cnaePf: string;
	aliquotaIssPf: number;
	descricaoServicoPf: string;
	ambiente: string;
}

interface NfseInvoice {
	id: string;
	amount: number;
	currency: string;
	status: string;
	nfseStatus: string | null;
	nfseNumber: string | null;
	nfseProtocol: string | null;
	nfseXmlUrl: string | null;
	nfsePdfUrl: string | null;
	nfseIssuedAt: string | null;
	nfseError: string | null;
	nfseCanceledAt: string | null;
	createdAt: string;
	updatedAt: string;
	user: {
		name: string;
		email: string;
		cpfCnpj: string | null;
	};
}

interface NfeDashboardProps {
	config: NfeConfig | null;
	invoices: NfseInvoice[];
	stats: Record<string, number>;
}

const STATUS_COLORS: Record<string, string> = {
	PENDING: "bg-chart-2/10 text-chart-2",
	PROCESSING: "bg-chart-3/10 text-chart-3",
	ISSUED: "bg-primary/10 text-primary",
	ERROR: "bg-destructive/10 text-destructive",
	CANCELLED: "bg-muted text-muted-foreground",
};

export function NfeDashboard({ config, invoices, stats }: NfeDashboardProps) {
	const [isPending, startTransition] = useTransition();
	const [actionResult, setActionResult] = useState<string | null>(null);

	function handleRetry(invoiceId: string) {
		startTransition(async () => {
			const result = await retryNfse(invoiceId);
			setActionResult(
				result.success
					? `NFS-e emitida com sucesso: ${result.data?.numero ?? result.data?.chaveAcesso ?? ""}`
					: `Erro: ${result.error}`,
			);
		});
	}

	function handleCancel(invoiceId: string) {
		if (!confirm("Tem certeza que deseja cancelar esta NFS-e?")) return;
		startTransition(async () => {
			const result = await cancelNfseAction(invoiceId);
			setActionResult(
				result.success
					? "NFS-e cancelada com sucesso"
					: `Erro: ${result.error}`,
			);
		});
	}

	function handleIssue(invoiceId: string) {
		startTransition(async () => {
			const result = await issueNfse(invoiceId);
			setActionResult(
				result.success
					? `NFS-e emitida: ${result.data?.numero ?? result.data?.chaveAcesso ?? ""}`
					: `Erro: ${result.error}`,
			);
		});
	}

	function handleCheckStatus(invoiceId: string) {
		startTransition(async () => {
			const result = await getNfseStatus(invoiceId);
			setActionResult(
				result
					? `Status: ${result.nfseStatus} | Numero: ${result.nfseNumber || "N/A"}`
					: "Invoice nao encontrada",
			);
		});
	}

	const certExpiresAt = config?.certificateExpiresAt
		? new Date(config.certificateExpiresAt)
		: null;
	const certExpired = certExpiresAt ? new Date() >= certExpiresAt : false;
	const certExpiresSoon = certExpiresAt
		? new Date() >=
			new Date(certExpiresAt.getTime() - 30 * 24 * 60 * 60 * 1000)
		: false;

	return (
		<div className="space-y-6 p-6">
			<h1 className="text-2xl font-semibold">NFS-e Nacional</h1>

			{actionResult && (
				<div
					className={`rounded-xl border p-4 ${actionResult.startsWith("Erro") ? "border-destructive/30 bg-destructive/10 text-destructive" : "border-primary/30 bg-primary/10 text-primary"}`}
				>
					<div className="flex items-center justify-between">
						<span className="text-sm">{actionResult}</span>
						<button
							type="button"
							onClick={() => setActionResult(null)}
							className="text-sm underline"
						>
							Fechar
						</button>
					</div>
				</div>
			)}

			{/* Config Card */}
			<div className="rounded-xl border bg-card p-6">
				<h2 className="mb-4 text-lg font-medium">Configuracao</h2>
				{!config ? (
					<p className="text-sm text-muted-foreground">
						Nenhuma configuracao NFS-e encontrada. Execute o script
						de setup.
					</p>
				) : (
					<div className="grid grid-cols-2 gap-4 text-sm md:grid-cols-3">
						<div>
							<span className="text-muted-foreground">
								Status
							</span>
							<p
								className={`font-medium ${config.active ? "text-primary" : "text-destructive"}`}
							>
								{config.active ? "Ativo" : "Inativo"}
							</p>
						</div>
						<div>
							<span className="text-muted-foreground">
								Ambiente
							</span>
							<p
								className={`font-medium ${config.ambiente === "PRODUCAO" ? "text-chart-2" : "text-chart-3"}`}
							>
								{config.ambiente}
							</p>
						</div>
						<div>
							<span className="text-muted-foreground">CNPJ</span>
							<p className="font-mono">{config.cnpj}</p>
						</div>
						<div>
							<span className="text-zinc-500">Municipio</span>
							<p>
								{config.codigoMunicipio} / {config.uf}
							</p>
						</div>
						<div>
							<span className="text-zinc-500">
								Certificado A1
							</span>
							<p
								className={`font-medium ${certExpired ? "text-red-600" : certExpiresSoon ? "text-yellow-600" : "text-green-600"}`}
							>
								{certExpired
									? "EXPIRADO"
									: certExpiresSoon
										? "Expira em breve"
										: "Valido"}
								{certExpiresAt && (
									<span className="ml-1 text-xs text-zinc-500">
										(ate{" "}
										{certExpiresAt.toLocaleDateString(
											"pt-BR",
										)}
										)
									</span>
								)}
							</p>
						</div>
						<div>
							<span className="text-zinc-500">PJ</span>
							<p className="text-xs">
								{config.codigoServico} / CNAE {config.cnae} /{" "}
								{(config.aliquotaIss * 100).toFixed(2)}%
							</p>
						</div>
						<div>
							<span className="text-zinc-500">PF</span>
							<p className="text-xs">
								{config.codigoServicoPf} / CNAE {config.cnaePf}{" "}
								/ {(config.aliquotaIssPf * 100).toFixed(2)}%
							</p>
						</div>
					</div>
				)}
			</div>

			{/* Stats */}
			<div className="grid grid-cols-2 gap-4 md:grid-cols-5">
				{(
					[
						"PENDING",
						"PROCESSING",
						"ISSUED",
						"ERROR",
						"CANCELLED",
					] as const
				).map((status) => (
					<div
						key={status}
						className="rounded-lg border bg-white p-4"
					>
						<p className="text-sm text-zinc-500">{status}</p>
						<p className="text-2xl font-semibold">
							{stats[status] || 0}
						</p>
					</div>
				))}
			</div>

			{/* Invoices Table */}
			<div className="rounded-lg border bg-white">
				<div className="border-b p-4">
					<h2 className="text-lg font-medium">NFS-e Recentes</h2>
				</div>
				<div className="overflow-x-auto">
					<table className="w-full text-sm">
						<thead>
							<tr className="border-b bg-zinc-50 text-left text-zinc-500">
								<th className="px-4 py-3">Usuario</th>
								<th className="px-4 py-3">Valor</th>
								<th className="px-4 py-3">Status</th>
								<th className="px-4 py-3">Numero</th>
								<th className="px-4 py-3">Erro</th>
								<th className="px-4 py-3">Data</th>
								<th className="px-4 py-3">Acoes</th>
							</tr>
						</thead>
						<tbody>
							{invoices.length === 0 ? (
								<tr>
									<td
										colSpan={7}
										className="px-4 py-8 text-center text-zinc-400"
									>
										Nenhuma NFS-e encontrada.
									</td>
								</tr>
							) : (
								invoices.map((inv) => (
									<tr
										key={inv.id}
										className="border-b hover:bg-zinc-50/50"
									>
										<td className="px-4 py-3">
											<div className="font-medium">
												{inv.user.name}
											</div>
											<div className="text-xs text-zinc-500">
												{inv.user.email}
											</div>
											{inv.user.cpfCnpj && (
												<div className="font-mono text-xs text-zinc-400">
													{inv.user.cpfCnpj}
												</div>
											)}
										</td>
										<td className="px-4 py-3 font-mono">
											{new Intl.NumberFormat("pt-BR", {
												style: "currency",
												currency: inv.currency,
											}).format(inv.amount)}
										</td>
										<td className="px-4 py-3">
											<span
												className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[inv.nfseStatus || ""] || "bg-zinc-100"}`}
											>
												{inv.nfseStatus}
											</span>
										</td>
										<td className="px-4 py-3 font-mono text-xs">
											{inv.nfseNumber || "-"}
										</td>
										<td className="max-w-xs truncate px-4 py-3 text-xs text-red-600">
											{inv.nfseError || "-"}
										</td>
										<td className="px-4 py-3 text-xs text-zinc-500">
											{inv.nfseIssuedAt
												? new Date(
														inv.nfseIssuedAt,
													).toLocaleDateString(
														"pt-BR",
													)
												: new Date(
														inv.updatedAt,
													).toLocaleDateString(
														"pt-BR",
													)}
										</td>
										<td className="px-4 py-3">
											<div className="flex gap-1">
												{inv.nfseStatus === "ERROR" && (
													<Button
														variant="outline"
														size="xs"
														onClick={() =>
															handleRetry(inv.id)
														}
														disabled={isPending}
													>
														Retentar
													</Button>
												)}
												{inv.nfseStatus ===
													"ISSUED" && (
													<Button
														variant="destructive"
														size="xs"
														onClick={() =>
															handleCancel(inv.id)
														}
														disabled={isPending}
													>
														Cancelar
													</Button>
												)}
												{inv.nfseStatus ===
													"PENDING" && (
													<Button
														variant="default"
														size="xs"
														onClick={() =>
															handleIssue(inv.id)
														}
														disabled={isPending}
													>
														Emitir
													</Button>
												)}
												{inv.nfsePdfUrl && (
													<a
														href={inv.nfsePdfUrl}
														target="_blank"
														rel="noopener noreferrer"
														className="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground hover:bg-muted/80 transition-colors"
													>
														PDF
													</a>
												)}
												{inv.nfseXmlUrl && (
													<a
														href={inv.nfseXmlUrl}
														target="_blank"
														rel="noopener noreferrer"
														className="rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground hover:bg-muted/80 transition-colors"
													>
														XML
													</a>
												)}
												<button
													type="button"
													onClick={() =>
														handleCheckStatus(
															inv.id,
														)
													}
													disabled={isPending}
													className="rounded bg-zinc-100 px-2 py-1 text-xs hover:bg-zinc-200 disabled:opacity-50"
												>
													Status
												</button>
											</div>
										</td>
									</tr>
								))
							)}
						</tbody>
					</table>
				</div>
			</div>
		</div>
	);
}
