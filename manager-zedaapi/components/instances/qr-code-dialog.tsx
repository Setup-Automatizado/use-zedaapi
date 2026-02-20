"use client";

import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Loader2, QrCode, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

interface QrCodeDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	instanceName: string;
	qrCode?: string;
	loading?: boolean;
	onRefresh?: () => void;
}

export function QrCodeDialog({
	open,
	onOpenChange,
	instanceName,
	qrCode,
	loading = false,
	onRefresh,
}: QrCodeDialogProps) {
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>Conectar WhatsApp</DialogTitle>
					<DialogDescription>
						Escaneie o QR Code com o WhatsApp de{" "}
						<strong>{instanceName}</strong> para conectar.
					</DialogDescription>
				</DialogHeader>
				<div className="flex flex-col items-center gap-4 py-4">
					{loading ? (
						<div className="flex size-64 items-center justify-center rounded-xl bg-muted">
							<Loader2 className="size-8 animate-spin text-muted-foreground" />
						</div>
					) : qrCode ? (
						<div className="rounded-xl border bg-white p-4">
							{/* QR code image will be rendered here */}
							<img
								src={`data:image/png;base64,${qrCode}`}
								alt="QR Code WhatsApp"
								className="size-56"
							/>
						</div>
					) : (
						<div className="flex size-64 flex-col items-center justify-center gap-3 rounded-xl bg-muted">
							<QrCode className="size-10 text-muted-foreground" />
							<p className="text-sm text-muted-foreground">
								QR Code indisponivel
							</p>
						</div>
					)}
					{onRefresh && (
						<Button
							variant="outline"
							size="sm"
							onClick={onRefresh}
							disabled={loading}
						>
							<RefreshCw className="size-4" />
							Gerar novo QR Code
						</Button>
					)}
					<p className="text-center text-xs text-muted-foreground">
						O QR Code expira em 60 segundos. Apos escanear, aguarde
						a conexao ser confirmada.
					</p>
				</div>
			</DialogContent>
		</Dialog>
	);
}
