/**
 * Token Display Component
 *
 * Displays authentication tokens with security features:
 * - Masked display by default (bullets)
 * - Toggle visibility (eye icon)
 * - Copy to clipboard functionality
 *
 * @example
 * ```tsx
 * <TokenDisplay
 *   label="Instance Token"
 *   token={instance.token}
 *   description="Token único desta instância"
 * />
 * ```
 */

"use client";

import { Check, Copy, Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export interface TokenDisplayProps {
	/** Label for the token field */
	label: string;

	/** The actual token value */
	token: string;

	/** Optional description text */
	description?: string;
}

export function TokenDisplay({ label, token, description }: TokenDisplayProps) {
	const [isVisible, setIsVisible] = useState(false);
	const [isCopied, setIsCopied] = useState(false);

	const handleCopy = async () => {
		try {
			await navigator.clipboard.writeText(token);
			setIsCopied(true);
			toast.success("Token copiado!", {
				description: `${label} foi copiado para a área de transferência.`,
			});

			// Reset icon after 2 seconds
			setTimeout(() => setIsCopied(false), 2000);
		} catch {
			toast.error("Erro ao copiar", {
				description: "Não foi possível copiar o token.",
			});
		}
	};

	const displayValue = isVisible ? token : "••••••••••••••••";

	return (
		<div className="space-y-2">
			<Label htmlFor={`token-${label}`}>{label}</Label>
			{description && (
				<p className="text-sm text-muted-foreground">{description}</p>
			)}
			<div className="flex gap-2">
				<Input
					id={`token-${label}`}
					type="text"
					value={displayValue}
					readOnly
					className="font-mono"
				/>
				<Button
					type="button"
					variant="outline"
					size="icon"
					onClick={() => setIsVisible(!isVisible)}
					title={isVisible ? "Ocultar token" : "Mostrar token"}
				>
					{isVisible ? (
						<EyeOff className="h-4 w-4" />
					) : (
						<Eye className="h-4 w-4" />
					)}
				</Button>
				<Button
					type="button"
					variant="outline"
					size="icon"
					onClick={handleCopy}
					title="Copiar token"
				>
					{isCopied ? (
						<Check className="h-4 w-4 text-green-600" />
					) : (
						<Copy className="h-4 w-4" />
					)}
				</Button>
			</div>
		</div>
	);
}
