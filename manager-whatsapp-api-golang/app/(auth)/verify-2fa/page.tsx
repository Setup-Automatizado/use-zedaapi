"use client";

import {
	ArrowLeft,
	KeyRound,
	Loader2,
	Mail,
	ShieldCheck,
	Smartphone,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { TotpInput } from "@/components/two-factor/totp-input";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { twoFactor } from "@/lib/auth-client";
import { cn } from "@/lib/utils";

type VerificationMode = "totp" | "email" | "backup";

export default function Verify2FAPage() {
	const router = useRouter();
	const [mode, setMode] = useState<VerificationMode>("totp");
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	// TOTP state
	const [totpCode, setTotpCode] = useState("");

	// Email OTP state
	const [emailSent, setEmailSent] = useState(false);
	const [emailCode, setEmailCode] = useState("");
	const [resendCooldown, setResendCooldown] = useState(0);

	// Backup code state
	const [backupCode, setBackupCode] = useState("");

	// Shared state
	const [trustDevice, setTrustDevice] = useState(false);

	// Resend cooldown timer
	useEffect(() => {
		if (resendCooldown > 0) {
			const timer = setTimeout(
				() => setResendCooldown(resendCooldown - 1),
				1000,
			);
			return () => clearTimeout(timer);
		}
	}, [resendCooldown]);

	// Reset state when changing modes
	const handleModeChange = (newMode: VerificationMode) => {
		if (newMode === mode) return;
		setMode(newMode);
		setError(null);
		setTotpCode("");
		setEmailCode("");
		setBackupCode("");
		// Don't reset emailSent to preserve the "code sent" state
	};

	// TOTP verification
	const handleTotpSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();

		if (totpCode.length !== 6) {
			setError("Please enter a valid 6-digit code");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const result = await twoFactor.verifyTotp({
				code: totpCode,
				trustDevice,
			});

			if (result.error) {
				setError(result.error.message || "Invalid code. Please try again.");
				return;
			}

			router.push("/dashboard");
			router.refresh();
		} catch {
			setError("An error occurred while verifying the code. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	// Send Email OTP
	const handleSendEmailOtp = async () => {
		setIsLoading(true);
		setError(null);

		try {
			const result = await twoFactor.sendOtp();

			if (result.error) {
				setError(
					result.error.message || "Failed to send code. Please try again.",
				);
				return;
			}

			setEmailSent(true);
			setResendCooldown(60);
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	// Verify Email OTP
	const handleEmailSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();

		if (emailCode.length !== 6) {
			setError("Please enter a valid 6-digit code");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const result = await twoFactor.verifyOtp({
				code: emailCode,
				trustDevice,
			});

			if (result.error) {
				setError(result.error.message || "Invalid code. Please try again.");
				return;
			}

			router.push("/dashboard");
			router.refresh();
		} catch {
			setError("An error occurred while verifying the code. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	// Backup code verification
	const handleBackupSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();

		if (!backupCode.trim()) {
			setError("Please enter a backup code");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const result = await twoFactor.verifyBackupCode({
				code: backupCode.trim(),
			});

			if (result.error) {
				setError(
					result.error.message || "Invalid backup code. Please try again.",
				);
				return;
			}

			router.push("/dashboard");
			router.refresh();
		} catch {
			setError("An error occurred while verifying the code. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	// Get icon for current mode
	const getModeIcon = () => {
		switch (mode) {
			case "totp":
				return <Smartphone className="h-8 w-8 text-primary" />;
			case "email":
				return <Mail className="h-8 w-8 text-primary" />;
			case "backup":
				return <KeyRound className="h-8 w-8 text-primary" />;
		}
	};

	// Get title for current mode
	const getModeTitle = () => {
		switch (mode) {
			case "totp":
				return "Two-factor authentication";
			case "email":
				return "Verify with email";
			case "backup":
				return "Use backup code";
		}
	};

	// Get description for current mode
	const getModeDescription = () => {
		switch (mode) {
			case "totp":
				return "Enter the 6-digit code from your authenticator app";
			case "email":
				return emailSent
					? "Enter the 6-digit code sent to your email"
					: "We'll send a 6-digit verification code to your registered email address";
			case "backup":
				return "Enter one of your backup codes to sign in";
		}
	};

	return (
		<div className="space-y-6">
			{/* Mobile Logo */}
			<div className="flex items-center justify-center gap-3 lg:hidden">
				<Image
					src="/android-chrome-96x96.png"
					alt="WhatsApp Manager"
					width={48}
					height={48}
					className="rounded-xl"
					priority
				/>
				<span className="text-xl font-bold text-foreground">
					WhatsApp Manager
				</span>
			</div>

			<Card className="border-0 shadow-none lg:border lg:shadow-sm">
				<CardHeader className="space-y-1 pb-6 text-center">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
						{getModeIcon()}
					</div>
					<CardTitle className="text-2xl font-bold tracking-tight">
						{getModeTitle()}
					</CardTitle>
					<CardDescription className="text-muted-foreground">
						{getModeDescription()}
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-6">
					{/* Mode Selector */}
					<div className="flex gap-1 p-1 bg-muted rounded-lg">
						<button
							type="button"
							onClick={() => handleModeChange("totp")}
							disabled={isLoading}
							className={cn(
								"flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors",
								mode === "totp"
									? "bg-background shadow-sm text-foreground"
									: "text-muted-foreground hover:text-foreground",
								isLoading && "opacity-50 cursor-not-allowed",
							)}
						>
							<Smartphone className="h-4 w-4" />
							App
						</button>
						<button
							type="button"
							onClick={() => handleModeChange("email")}
							disabled={isLoading}
							className={cn(
								"flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors",
								mode === "email"
									? "bg-background shadow-sm text-foreground"
									: "text-muted-foreground hover:text-foreground",
								isLoading && "opacity-50 cursor-not-allowed",
							)}
						>
							<Mail className="h-4 w-4" />
							Email
						</button>
						<button
							type="button"
							onClick={() => handleModeChange("backup")}
							disabled={isLoading}
							className={cn(
								"flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors",
								mode === "backup"
									? "bg-background shadow-sm text-foreground"
									: "text-muted-foreground hover:text-foreground",
								isLoading && "opacity-50 cursor-not-allowed",
							)}
						>
							<KeyRound className="h-4 w-4" />
							Backup
						</button>
					</div>

					{/* Error Message */}
					{error && (
						<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive">
							{error}
						</div>
					)}

					{/* TOTP Mode */}
					{mode === "totp" && (
						<form onSubmit={handleTotpSubmit} className="space-y-6">
							<div className="space-y-4">
								<TotpInput
									value={totpCode}
									onChange={setTotpCode}
									disabled={isLoading}
									autoFocus
								/>

								<div className="flex items-center gap-3">
									<Checkbox
										id="trust-device-totp"
										checked={trustDevice}
										onCheckedChange={(checked) =>
											setTrustDevice(checked === true)
										}
										disabled={isLoading}
									/>
									<Label
										htmlFor="trust-device-totp"
										className="text-sm cursor-pointer"
									>
										Trust this device for 30 days
									</Label>
								</div>
							</div>

							<Button
								type="submit"
								className="w-full h-11 font-medium"
								disabled={isLoading || totpCode.length !== 6}
							>
								{isLoading ? (
									<>
										<Loader2 className="mr-2 h-4 w-4 animate-spin" />
										Verifying...
									</>
								) : (
									<>
										<ShieldCheck className="mr-2 h-4 w-4" />
										Verify code
									</>
								)}
							</Button>
						</form>
					)}

					{/* Email Mode */}
					{mode === "email" && (
						<>
							{!emailSent ? (
								<div className="space-y-6">
									<div className="text-center space-y-2">
										<div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
											<Mail className="h-6 w-6 text-primary" />
										</div>
										<p className="text-sm text-muted-foreground">
											Click the button below to receive a verification code at
											your registered email address.
										</p>
									</div>

									<Button
										type="button"
										onClick={handleSendEmailOtp}
										className="w-full h-11 font-medium"
										disabled={isLoading}
									>
										{isLoading ? (
											<>
												<Loader2 className="mr-2 h-4 w-4 animate-spin" />
												Sending...
											</>
										) : (
											<>
												<Mail className="mr-2 h-4 w-4" />
												Send verification code
											</>
										)}
									</Button>
								</div>
							) : (
								<form onSubmit={handleEmailSubmit} className="space-y-6">
									<div className="space-y-4">
										<TotpInput
											value={emailCode}
											onChange={setEmailCode}
											disabled={isLoading}
											autoFocus
										/>

										<div className="flex items-center gap-3">
											<Checkbox
												id="trust-device-email"
												checked={trustDevice}
												onCheckedChange={(checked) =>
													setTrustDevice(checked === true)
												}
												disabled={isLoading}
											/>
											<Label
												htmlFor="trust-device-email"
												className="text-sm cursor-pointer"
											>
												Trust this device for 30 days
											</Label>
										</div>
									</div>

									<Button
										type="submit"
										className="w-full h-11 font-medium"
										disabled={isLoading || emailCode.length !== 6}
									>
										{isLoading ? (
											<>
												<Loader2 className="mr-2 h-4 w-4 animate-spin" />
												Verifying...
											</>
										) : (
											<>
												<ShieldCheck className="mr-2 h-4 w-4" />
												Verify code
											</>
										)}
									</Button>

									<div className="text-center">
										<button
											type="button"
											onClick={handleSendEmailOtp}
											disabled={resendCooldown > 0 || isLoading}
											className="text-sm text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
										>
											{resendCooldown > 0
												? `Resend code in ${resendCooldown}s`
												: "Resend code"}
										</button>
									</div>
								</form>
							)}
						</>
					)}

					{/* Backup Mode */}
					{mode === "backup" && (
						<form onSubmit={handleBackupSubmit} className="space-y-6">
							<div className="space-y-2">
								<Label htmlFor="backup-code">Backup code</Label>
								<Input
									id="backup-code"
									name="backup-code"
									type="text"
									placeholder="Enter your backup code"
									value={backupCode}
									onChange={(e) => {
										setBackupCode(e.target.value);
										setError(null);
									}}
									disabled={isLoading}
									required
									autoComplete="off"
									className="h-11 font-mono"
								/>
							</div>

							<Button
								type="submit"
								className="w-full h-11 font-medium"
								disabled={isLoading || !backupCode.trim()}
							>
								{isLoading ? (
									<>
										<Loader2 className="mr-2 h-4 w-4 animate-spin" />
										Verifying...
									</>
								) : (
									<>
										<KeyRound className="mr-2 h-4 w-4" />
										Verify backup code
									</>
								)}
							</Button>
						</form>
					)}

					{/* Back to Login */}
					<div className="pt-2">
						<Link href="/login" className="block">
							<Button
								variant="ghost"
								className="w-full h-11"
								disabled={isLoading}
							>
								<ArrowLeft className="mr-2 h-4 w-4" />
								Back to login
							</Button>
						</Link>
					</div>

					{/* Help Link */}
					<div className="text-center">
						<Link
							href="/forgot-password"
							className="text-sm text-muted-foreground hover:text-foreground transition-colors"
						>
							Lost access to your authentication methods?
						</Link>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
