"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";

interface ReferralLinkProps {
	code: string;
	baseUrl?: string;
}

export function ReferralLink({
	code,
	baseUrl = process.env.NEXT_PUBLIC_APP_URL || "https://manager.zedaapi.com",
}: ReferralLinkProps) {
	const [copied, setCopied] = useState(false);

	const referralUrl = `${baseUrl}/sign-up?ref=${code}&utm_source=affiliate&utm_medium=referral&utm_campaign=${code}`;

	async function handleCopy() {
		try {
			await navigator.clipboard.writeText(referralUrl);
		} catch {
			const input = document.createElement("input");
			input.value = referralUrl;
			document.body.appendChild(input);
			input.select();
			document.execCommand("copy");
			document.body.removeChild(input);
		}
		setCopied(true);
		toast.success("Link copiado para a area de transferencia");
		setTimeout(() => setCopied(false), 2000);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm font-medium">
					Seu link de indicacao
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div className="flex gap-2">
					<Input
						readOnly
						value={referralUrl}
						className="font-mono text-xs"
						onClick={(e) => e.currentTarget.select()}
					/>
					<Button
						variant="outline"
						onClick={handleCopy}
						className="shrink-0 gap-1.5"
					>
						{copied ? (
							<Check className="size-4 text-primary" />
						) : (
							<Copy className="size-4" />
						)}
						{copied ? "Copiado" : "Copiar"}
					</Button>
				</div>
				<p className="mt-2 text-xs text-muted-foreground">
					Compartilhe este link para receber comissoes por cada
					indicacao convertida.
				</p>
			</CardContent>
		</Card>
	);
}
