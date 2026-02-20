"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { InvoiceTable } from "@/components/billing/invoice-table";
import {
	requestSicrediPayment,
	createBillingPortal,
} from "@/server/actions/billing";
import { ExternalLink } from "lucide-react";

interface BillingClientProps {
	initialInvoices: Array<{
		id: string;
		amount: number;
		currency: string;
		status: string;
		paidAt: string | null;
		dueDate: string | null;
		paymentMethod: string | null;
		pdfUrl: string | null;
		createdAt: string;
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
	}>;
	total: number;
	pageSize: number;
}

export function BillingClient({
	initialInvoices,
	total,
	pageSize,
}: BillingClientProps) {
	const router = useRouter();
	const [portalLoading, setPortalLoading] = useState(false);

	async function handleRequestPayment(
		invoiceId: string,
		type: "pix" | "boleto_hibrido",
	) {
		try {
			const result = await requestSicrediPayment(invoiceId, type);
			// Refresh to show the new charge data
			router.refresh();
		} catch (err) {
			console.error("Payment request error:", err);
		}
	}

	async function handleBillingPortal() {
		setPortalLoading(true);
		try {
			const result = await createBillingPortal();
			if (result.success && result.data?.url) {
				window.location.href = result.data.url;
			}
		} catch (err) {
			console.error("Portal error:", err);
		} finally {
			setPortalLoading(false);
		}
	}

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div />
				<Button
					variant="outline"
					size="sm"
					onClick={handleBillingPortal}
					disabled={portalLoading}
					className="gap-1.5"
				>
					<ExternalLink className="size-3.5" />
					{portalLoading ? "Abrindo..." : "Portal de pagamento"}
				</Button>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">Faturas</CardTitle>
				</CardHeader>
				<CardContent>
					<InvoiceTable
						invoices={initialInvoices}
						onRequestPayment={handleRequestPayment}
					/>
				</CardContent>
			</Card>

			{total > pageSize && (
				<div className="flex justify-center">
					<p className="text-sm text-muted-foreground">
						Mostrando {initialInvoices.length} de {total} faturas
					</p>
				</div>
			)}
		</div>
	);
}
