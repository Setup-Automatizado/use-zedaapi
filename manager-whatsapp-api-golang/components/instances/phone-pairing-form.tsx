"use client";

import { CheckCircle2, Copy, Loader2, Phone } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Field,
	FieldContent,
	FieldDescription,
	FieldError,
	FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { formatPhoneInput } from "@/lib/phone";

export interface PhonePairingFormProps {
	instanceId: string;
	instanceToken: string;
	className?: string;
}

export function PhonePairingForm({
	instanceId,
	// instanceToken is kept in the interface for API consistency
	instanceToken: _instanceToken,
	className,
}: PhonePairingFormProps) {
	void _instanceToken; // Mark as intentionally unused
	const [phone, setPhone] = useState("");
	const [pairingCode, setPairingCode] = useState<string | null>(null);
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [copied, setCopied] = useState(false);

	const handlePhoneChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const formatted = formatPhoneInput(e.target.value);
		setPhone(formatted);
		setError(null);
	};

	const handleGenerateCode = async () => {
		// Extract only digits
		const phoneDigits = phone.replace(/\D/g, "");

		// Validate phone number (Brazilian format: 11 digits including DDD)
		if (phoneDigits.length < 10 || phoneDigits.length > 11) {
			setError("Invalid phone number. Use format (XX) XXXXX-XXXX");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const response = await fetch(
				`/api/instances/${instanceId}/phone-code/${phoneDigits}`,
				{
					method: "GET",
					headers: {
						"Content-Type": "application/json",
					},
				},
			);

			if (!response.ok) {
				throw new Error("Error generating pairing code");
			}

			const data = await response.json();
			setPairingCode(data.code);
			toast.success("Code generated successfully!");
		} catch (error) {
			const message =
				error instanceof Error ? error.message : "Error generating code";
			setError(message);
			toast.error(message);
		} finally {
			setIsLoading(false);
		}
	};

	const handleCopyCode = async () => {
		if (!pairingCode) return;

		try {
			await navigator.clipboard.writeText(pairingCode);
			setCopied(true);
			toast.success("Code copied!");
			setTimeout(() => setCopied(false), 2000);
		} catch {
			toast.error("Error copying code");
		}
	};

	return (
		<Card className={className}>
			<CardHeader>
				<CardTitle>Phone Pairing</CardTitle>
				<CardDescription>
					Use a 6-digit code to connect without scanning QR Code
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<Field>
					<FieldLabel htmlFor="phone">Phone Number</FieldLabel>
					<FieldContent>
						<div className="flex gap-2">
							<div className="relative flex-1">
								<Phone className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
								<Input
									id="phone"
									type="tel"
									placeholder="(11) 99999-9999"
									value={phone}
									onChange={handlePhoneChange}
									disabled={isLoading}
									className="pl-10"
									maxLength={15}
								/>
							</div>
							<Button
								onClick={handleGenerateCode}
								disabled={isLoading || phone.length < 14}
							>
								{isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
								{isLoading ? "Generating..." : "Generate Code"}
							</Button>
						</div>
						<FieldDescription>
							Enter the WhatsApp number you want to connect
						</FieldDescription>
						{error && <FieldError>{error}</FieldError>}
					</FieldContent>
				</Field>

				{pairingCode && (
					<div className="rounded-lg border bg-muted/50 p-6">
						<div className="mb-4 text-center">
							<p className="mb-2 text-sm font-medium text-muted-foreground">
								Your pairing code:
							</p>
							<div className="mb-3 font-mono text-4xl font-bold tracking-widest">
								{pairingCode}
							</div>
							<Button
								onClick={handleCopyCode}
								variant="outline"
								size="sm"
								className="gap-2"
							>
								{copied ? (
									<>
										<CheckCircle2 className="h-4 w-4 text-green-600" />
										Copied!
									</>
								) : (
									<>
										<Copy className="h-4 w-4" />
										Copy Code
									</>
								)}
							</Button>
						</div>

						<div className="space-y-2 text-sm text-muted-foreground">
							<p className="font-medium">How to use the code:</p>
							<ol className="list-inside list-decimal space-y-1">
								<li>Open WhatsApp on your phone</li>
								<li>Tap Menu (3 dots) and select &quot;Linked devices&quot;</li>
								<li>Tap &quot;Link a device&quot;</li>
								<li>Select &quot;Link with phone number instead&quot;</li>
								<li>Enter the code shown above</li>
							</ol>
						</div>
					</div>
				)}

				{!pairingCode && (
					<div className="rounded-md bg-blue-50 p-4 dark:bg-blue-950/20">
						<p className="text-sm text-blue-900 dark:text-blue-100">
							<strong>Tip:</strong> This option is ideal if you cannot scan the
							QR Code or prefer a faster connection.
						</p>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
