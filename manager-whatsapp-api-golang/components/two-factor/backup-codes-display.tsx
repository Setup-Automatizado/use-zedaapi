"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Copy, Download, CheckCircle2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface BackupCodesDisplayProps {
	codes: string[];
	onConfirm?: () => void;
	showConfirmation?: boolean;
	className?: string;
}

/**
 * Backup Codes Display Component
 *
 * Displays backup codes in a grid with copy and download functionality.
 * Optionally shows a confirmation checkbox for user acknowledgment.
 */
export function BackupCodesDisplay({
	codes,
	onConfirm,
	showConfirmation = true,
	className,
}: BackupCodesDisplayProps) {
	const [copied, setCopied] = React.useState(false);
	const [confirmed, setConfirmed] = React.useState(false);

	const handleCopyAll = async () => {
		const codesText = codes.join("\n");
		await navigator.clipboard.writeText(codesText);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};

	const handleDownload = () => {
		const codesText = [
			"WhatsApp Manager - Backup Codes",
			"================================",
			"",
			"Keep these codes in a safe place.",
			"Each code can only be used once.",
			"",
			...codes.map((code, index) => `${index + 1}. ${code}`),
			"",
			`Generated: ${new Date().toISOString()}`,
		].join("\n");

		const blob = new Blob([codesText], { type: "text/plain" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "whatsapp-manager-backup-codes.txt";
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	};

	const handleConfirmChange = (checked: boolean) => {
		setConfirmed(checked);
		if (checked && onConfirm) {
			onConfirm();
		}
	};

	return (
		<div className={cn("space-y-4", className)}>
			{/* Codes Grid */}
			<div className="grid grid-cols-2 gap-2 p-4 bg-muted/50 rounded-xl">
				{codes.map((code, index) => (
					<div
						key={index}
						className="font-mono text-sm p-2 bg-background rounded-lg text-center"
					>
						{code}
					</div>
				))}
			</div>

			{/* Action Buttons */}
			<div className="flex gap-2">
				<Button
					type="button"
					variant="outline"
					onClick={handleCopyAll}
					className="flex-1 h-10"
				>
					{copied ? (
						<>
							<CheckCircle2 className="mr-2 h-4 w-4 text-primary" />
							Copied!
						</>
					) : (
						<>
							<Copy className="mr-2 h-4 w-4" />
							Copy all
						</>
					)}
				</Button>
				<Button
					type="button"
					variant="outline"
					onClick={handleDownload}
					className="flex-1 h-10"
				>
					<Download className="mr-2 h-4 w-4" />
					Download
				</Button>
			</div>

			{/* Confirmation Checkbox */}
			{showConfirmation && (
				<div className="flex items-start gap-3 p-3 bg-primary/5 rounded-lg">
					<Checkbox
						id="confirm-backup"
						checked={confirmed}
						onCheckedChange={handleConfirmChange}
						className="mt-0.5"
					/>
					<Label
						htmlFor="confirm-backup"
						className="text-sm leading-relaxed cursor-pointer"
					>
						I have saved these backup codes in a secure location. I
						understand that each code can only be used once and I
						will need them if I lose access to my authenticator app.
					</Label>
				</div>
			)}
		</div>
	);
}

export default BackupCodesDisplay;
