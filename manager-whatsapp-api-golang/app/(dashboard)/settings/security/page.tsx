"use client";

import { useState, useEffect } from "react";
import { useSession, twoFactor } from "@/lib/auth-client";
import QRCode from "react-qr-code";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { TotpInput } from "@/components/two-factor/totp-input";
import { BackupCodesDisplay } from "@/components/two-factor/backup-codes-display";
import {
	Loader2,
	ShieldCheck,
	ShieldOff,
	KeyRound,
	RefreshCw,
	CheckCircle2,
	AlertTriangle,
	Smartphone,
	Mail,
} from "lucide-react";

type TwoFactorMethod = "totp" | "email";

type SetupStep =
	| "idle"
	| "choose-method"
	| "password"
	| "qrcode"
	| "verify"
	| "email-sent"
	| "backup"
	| "success"
	| "disabling";

export default function SecuritySettingsPage() {
	const { data: session, isPending: isSessionLoading } = useSession();
	const [setupStep, setSetupStep] = useState<SetupStep>("idle");
	const [selectedMethod, setSelectedMethod] =
		useState<TwoFactorMethod | null>(null);
	const [currentMethod, setCurrentMethod] = useState<TwoFactorMethod | null>(
		null,
	);
	const [isLoading, setIsLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const [password, setPassword] = useState("");
	const [totpUri, setTotpUri] = useState("");
	const [verificationCode, setVerificationCode] = useState("");
	const [backupCodes, setBackupCodes] = useState<string[]>([]);
	const [backupConfirmed, setBackupConfirmed] = useState(false);
	const [isRegenerating, setIsRegenerating] = useState(false);

	const isTwoFactorEnabled = session?.user?.twoFactorEnabled ?? false;

	// Fetch current 2FA method when component mounts
	useEffect(() => {
		if (isTwoFactorEnabled) {
			fetch("/api/two-factor")
				.then((res) => res.json())
				.then((data) => {
					if (data.method) {
						setCurrentMethod(data.method as TwoFactorMethod);
					}
				})
				.catch(console.error);
		}
	}, [isTwoFactorEnabled]);

	const resetState = () => {
		setSetupStep("idle");
		setSelectedMethod(null);
		setPassword("");
		setTotpUri("");
		setVerificationCode("");
		setBackupCodes([]);
		setBackupConfirmed(false);
		setIsRegenerating(false);
		setError(null);
	};

	const handleStartEnable = () => {
		setSetupStep("choose-method");
		setError(null);
	};

	const handleSelectMethod = (method: TwoFactorMethod) => {
		setSelectedMethod(method);
		setSetupStep("password");
	};

	const handlePasswordSubmit = async (
		e: React.FormEvent<HTMLFormElement>,
	) => {
		e.preventDefault();

		if (!password) {
			setError("Please enter your password");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			// If regenerating backup codes, just get new codes and show them
			if (isRegenerating) {
				const result = await twoFactor.enable({
					password,
				});

				if (result.error) {
					setError(
						result.error.message ||
							"Failed to regenerate backup codes. Please check your password.",
					);
					return;
				}

				if (result.data?.backupCodes) {
					setBackupCodes(result.data.backupCodes);
					setSetupStep("backup");
				}
				return;
			}

			if (selectedMethod === "totp") {
				// Enable TOTP and get QR code
				const result = await twoFactor.enable({
					password,
				});

				if (result.error) {
					setError(
						result.error.message ||
							"Failed to generate QR code. Please check your password.",
					);
					return;
				}

				if (result.data?.totpURI) {
					setTotpUri(result.data.totpURI);
					if (result.data?.backupCodes) {
						setBackupCodes(result.data.backupCodes);
					}
					setSetupStep("qrcode");
				}
			} else if (selectedMethod === "email") {
				// First enable 2FA (this creates the TwoFactor record and triggers email notification)
				const enableResult = await twoFactor.enable({
					password,
				});

				if (enableResult.error) {
					setError(
						enableResult.error.message ||
							"Failed to enable 2FA. Please check your password.",
					);
					return;
				}

				// Store backup codes if returned
				if (enableResult.data?.backupCodes) {
					setBackupCodes(enableResult.data.backupCodes);
				}

				// Send OTP via email for verification
				const result = await twoFactor.sendOtp();

				if (result.error) {
					setError(
						result.error.message ||
							"Failed to send verification code. Please try again.",
					);
					return;
				}

				setSetupStep("email-sent");
			}
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleQrCodeNext = () => {
		setSetupStep("verify");
		setVerificationCode("");
	};

	const handleVerifyTotpCode = async (
		e: React.FormEvent<HTMLFormElement>,
	) => {
		e.preventDefault();

		if (verificationCode.length !== 6) {
			setError("Please enter a valid 6-digit code");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const verifyResult = await twoFactor.verifyTotp({
				code: verificationCode,
			});

			if (verifyResult.error) {
				setError(
					verifyResult.error.message ||
						"Invalid code. Please try again.",
				);
				return;
			}

			setSetupStep("backup");
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleVerifyEmailOtp = async (
		e: React.FormEvent<HTMLFormElement>,
	) => {
		e.preventDefault();

		if (verificationCode.length !== 6) {
			setError("Please enter a valid 6-digit code");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const verifyResult = await twoFactor.verifyOtp({
				code: verificationCode,
			});

			if (verifyResult.error) {
				setError(
					verifyResult.error.message ||
						"Invalid code. Please try again.",
				);
				return;
			}

			// After verifying email OTP, show backup codes if available
			if (backupCodes.length > 0) {
				setSetupStep("backup");
			} else {
				// Save the method to the database before completing
				if (selectedMethod) {
					try {
						await fetch("/api/two-factor", {
							method: "PATCH",
							headers: { "Content-Type": "application/json" },
							body: JSON.stringify({ method: selectedMethod }),
						});
					} catch (error) {
						console.error("Failed to save 2FA method:", error);
					}
				}
				setSetupStep("success");
				setTimeout(() => {
					window.location.reload();
				}, 1500);
			}
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleResendEmailOtp = async () => {
		setIsLoading(true);
		setError(null);

		try {
			const result = await twoFactor.sendOtp();

			if (result.error) {
				setError(result.error.message || "Failed to resend code.");
				return;
			}

			setError(null);
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleBackupConfirm = () => {
		setBackupConfirmed(true);
	};

	const handleSetupComplete = async () => {
		// Save the method to the database
		if (selectedMethod) {
			try {
				await fetch("/api/two-factor", {
					method: "PATCH",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ method: selectedMethod }),
				});
			} catch (error) {
				console.error("Failed to save 2FA method:", error);
			}
		}
		resetState();
		window.location.reload();
	};

	const handleStartDisable = () => {
		setSetupStep("disabling");
		setPassword("");
		setError(null);
	};

	const handleDisable = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();

		if (!password) {
			setError("Please enter your password");
			return;
		}

		setIsLoading(true);
		setError(null);

		try {
			const result = await twoFactor.disable({
				password,
			});

			if (result.error) {
				setError(
					result.error.message ||
						"Failed to disable two-factor authentication.",
				);
				return;
			}

			resetState();
			window.location.reload();
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleRegenerateBackupCodes = async () => {
		setSetupStep("password");
		// Use the current method, default to totp if not set
		setSelectedMethod(currentMethod || "totp");
		setIsRegenerating(true);
	};

	if (isSessionLoading) {
		return (
			<div className="flex items-center justify-center min-h-[400px]">
				<Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-3xl font-bold tracking-tight">
					Security Settings
				</h1>
				<p className="text-muted-foreground">
					Manage your account security and two-factor authentication
				</p>
			</div>

			{/* Two-Factor Authentication Section */}
			<Card>
				<CardHeader>
					<div className="flex items-center gap-3">
						{isTwoFactorEnabled ? (
							<div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
								<ShieldCheck className="h-5 w-5 text-primary" />
							</div>
						) : (
							<div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
								<ShieldOff className="h-5 w-5 text-muted-foreground" />
							</div>
						)}
						<div>
							<CardTitle>Two-Factor Authentication</CardTitle>
							<CardDescription>
								{isTwoFactorEnabled
									? "Two-factor authentication is enabled"
									: "Add an extra layer of security to your account"}
							</CardDescription>
						</div>
					</div>
				</CardHeader>
				<CardContent className="space-y-6">
					{error && (
						<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive flex items-start gap-2">
							<AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
							{error}
						</div>
					)}

					{/* Idle State */}
					{setupStep === "idle" && (
						<div className="space-y-4">
							{isTwoFactorEnabled ? (
								<div className="space-y-3">
									<div className="flex items-center gap-2 p-3 bg-primary/5 rounded-lg">
										<CheckCircle2 className="h-5 w-5 text-primary" />
										<span className="text-sm">
											Your account is protected with
											two-factor authentication
										</span>
									</div>
									{currentMethod && (
										<div className="flex items-center gap-2 p-3 bg-muted/50 rounded-lg">
											{currentMethod === "totp" ? (
												<>
													<Smartphone className="h-4 w-4 text-muted-foreground" />
													<span className="text-sm text-muted-foreground">
														Method:{" "}
														<strong className="text-foreground">
															Authenticator App
														</strong>
													</span>
												</>
											) : (
												<>
													<Mail className="h-4 w-4 text-muted-foreground" />
													<span className="text-sm text-muted-foreground">
														Method:{" "}
														<strong className="text-foreground">
															Email Code
														</strong>
													</span>
												</>
											)}
										</div>
									)}
								</div>
							) : (
								<p className="text-sm text-muted-foreground">
									Two-factor authentication adds an extra
									layer of security to your account. Choose
									between an authenticator app or email
									verification.
								</p>
							)}

							<div className="flex gap-3">
								{isTwoFactorEnabled ? (
									<>
										<Button
											variant="outline"
											onClick={handleStartDisable}
										>
											<ShieldOff className="mr-2 h-4 w-4" />
											Disable 2FA
										</Button>
										<Button
											variant="outline"
											onClick={
												handleRegenerateBackupCodes
											}
										>
											<RefreshCw className="mr-2 h-4 w-4" />
											Regenerate backup codes
										</Button>
									</>
								) : (
									<Button onClick={handleStartEnable}>
										<ShieldCheck className="mr-2 h-4 w-4" />
										Enable 2FA
									</Button>
								)}
							</div>
						</div>
					)}

					{/* Choose Method Step */}
					{setupStep === "choose-method" && (
						<div className="space-y-4">
							<p className="text-sm text-muted-foreground">
								Choose your preferred two-factor authentication
								method:
							</p>

							<div className="grid gap-4 sm:grid-cols-2">
								<button
									type="button"
									onClick={() => handleSelectMethod("totp")}
									className="flex flex-col items-center gap-3 p-6 rounded-lg border-2 border-muted hover:border-primary hover:bg-primary/5 transition-colors text-left"
								>
									<div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
										<Smartphone className="h-6 w-6 text-primary" />
									</div>
									<div className="text-center">
										<h3 className="font-semibold">
											Authenticator App
										</h3>
										<p className="text-sm text-muted-foreground mt-1">
											Use Google Authenticator, Authy, or
											similar apps
										</p>
									</div>
								</button>

								<button
									type="button"
									onClick={() => handleSelectMethod("email")}
									className="flex flex-col items-center gap-3 p-6 rounded-lg border-2 border-muted hover:border-primary hover:bg-primary/5 transition-colors text-left"
								>
									<div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
										<Mail className="h-6 w-6 text-primary" />
									</div>
									<div className="text-center">
										<h3 className="font-semibold">
											Email Code
										</h3>
										<p className="text-sm text-muted-foreground mt-1">
											Receive a verification code via
											email
										</p>
									</div>
								</button>
							</div>

							<Button variant="outline" onClick={resetState}>
								Cancel
							</Button>
						</div>
					)}

					{/* Password Confirmation Step */}
					{setupStep === "password" && (
						<form
							onSubmit={handlePasswordSubmit}
							className="space-y-4"
						>
							<div className="flex items-center gap-2 p-3 bg-muted/50 rounded-lg">
								{isRegenerating ? (
									<>
										<RefreshCw className="h-4 w-4" />
										<span className="text-sm">
											Regenerate Backup Codes
										</span>
									</>
								) : selectedMethod === "totp" ? (
									<>
										<Smartphone className="h-4 w-4" />
										<span className="text-sm">
											Authenticator App
										</span>
									</>
								) : (
									<>
										<Mail className="h-4 w-4" />
										<span className="text-sm">
											Email Code
										</span>
									</>
								)}
							</div>

							<div className="space-y-2">
								<Label htmlFor="password">
									Confirm your password
								</Label>
								<Input
									id="password"
									name="password"
									type="password"
									placeholder="Enter your password"
									value={password}
									onChange={(e) => {
										setPassword(e.target.value);
										setError(null);
									}}
									disabled={isLoading}
									required
									autoComplete="current-password"
									className="h-11"
								/>
							</div>

							<div className="flex gap-3">
								<Button
									type="button"
									variant="outline"
									onClick={() =>
										isRegenerating
											? resetState()
											: setSetupStep("choose-method")
									}
									disabled={isLoading}
								>
									{isRegenerating ? "Cancel" : "Back"}
								</Button>
								<Button
									type="submit"
									disabled={isLoading || !password}
								>
									{isLoading ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											{isRegenerating
												? "Regenerating..."
												: selectedMethod === "email"
													? "Sending..."
													: "Generating..."}
										</>
									) : (
										"Continue"
									)}
								</Button>
							</div>
						</form>
					)}

					{/* QR Code Step (TOTP only) */}
					{setupStep === "qrcode" && totpUri && (
						<div className="space-y-6">
							<div className="space-y-4">
								<p className="text-sm text-muted-foreground">
									Scan this QR code with your authenticator
									app (such as Google Authenticator, Authy, or
									1Password).
								</p>

								<div className="flex justify-center p-6 bg-white rounded-xl">
									<QRCode value={totpUri} size={200} />
								</div>

								<details className="text-sm">
									<summary className="cursor-pointer text-muted-foreground hover:text-foreground">
										Can&apos;t scan the code? Enter manually
									</summary>
									<div className="mt-2 p-3 bg-muted rounded-lg">
										<code className="text-xs break-all">
											{totpUri}
										</code>
									</div>
								</details>
							</div>

							<div className="flex gap-3">
								<Button
									type="button"
									variant="outline"
									onClick={resetState}
								>
									Cancel
								</Button>
								<Button onClick={handleQrCodeNext}>
									I&apos;ve scanned the code
								</Button>
							</div>
						</div>
					)}

					{/* Verify TOTP Code Step */}
					{setupStep === "verify" && (
						<form
							onSubmit={handleVerifyTotpCode}
							className="space-y-6"
						>
							<div className="space-y-4">
								<p className="text-sm text-muted-foreground">
									Enter the 6-digit code from your
									authenticator app to verify the setup.
								</p>

								<TotpInput
									value={verificationCode}
									onChange={setVerificationCode}
									disabled={isLoading}
									autoFocus
								/>
							</div>

							<div className="flex gap-3">
								<Button
									type="button"
									variant="outline"
									onClick={() => setSetupStep("qrcode")}
									disabled={isLoading}
								>
									Back
								</Button>
								<Button
									type="submit"
									disabled={
										isLoading ||
										verificationCode.length !== 6
									}
								>
									{isLoading ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Verifying...
										</>
									) : (
										"Verify and enable"
									)}
								</Button>
							</div>
						</form>
					)}

					{/* Email OTP Sent Step */}
					{setupStep === "email-sent" && (
						<form
							onSubmit={handleVerifyEmailOtp}
							className="space-y-6"
						>
							<div className="space-y-4">
								<div className="flex items-center gap-2 p-3 bg-primary/5 rounded-lg">
									<Mail className="h-5 w-5 text-primary" />
									<span className="text-sm">
										We sent a verification code to{" "}
										<strong>{session?.user?.email}</strong>
									</span>
								</div>

								<p className="text-sm text-muted-foreground">
									Enter the 6-digit code from your email to
									enable two-factor authentication.
								</p>

								<TotpInput
									value={verificationCode}
									onChange={setVerificationCode}
									disabled={isLoading}
									autoFocus
								/>

								<button
									type="button"
									onClick={handleResendEmailOtp}
									disabled={isLoading}
									className="text-sm text-primary hover:underline disabled:opacity-50"
								>
									Didn&apos;t receive the code? Resend
								</button>
							</div>

							<div className="flex gap-3">
								<Button
									type="button"
									variant="outline"
									onClick={resetState}
									disabled={isLoading}
								>
									Cancel
								</Button>
								<Button
									type="submit"
									disabled={
										isLoading ||
										verificationCode.length !== 6
									}
								>
									{isLoading ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Verifying...
										</>
									) : (
										"Verify and enable"
									)}
								</Button>
							</div>
						</form>
					)}

					{/* Backup Codes Step */}
					{setupStep === "backup" && backupCodes.length > 0 && (
						<div className="space-y-6">
							<div className="space-y-2">
								<div className="flex items-center gap-2">
									<KeyRound className="h-5 w-5 text-primary" />
									<h3 className="font-semibold">
										Backup Codes
									</h3>
								</div>
								<p className="text-sm text-muted-foreground">
									Save these backup codes in a secure
									location. You can use them to sign in if you
									lose access to your authenticator app. Each
									code can only be used once.
								</p>
							</div>

							<BackupCodesDisplay
								codes={backupCodes}
								onConfirm={handleBackupConfirm}
								showConfirmation={!isTwoFactorEnabled}
							/>

							<Button
								onClick={handleSetupComplete}
								disabled={
									!backupConfirmed && !isTwoFactorEnabled
								}
								className="w-full"
							>
								<CheckCircle2 className="mr-2 h-4 w-4" />
								{isTwoFactorEnabled ? "Done" : "Complete setup"}
							</Button>
						</div>
					)}

					{/* Success Step */}
					{setupStep === "success" && (
						<div className="flex flex-col items-center gap-4 py-6">
							<div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
								<CheckCircle2 className="h-8 w-8 text-primary" />
							</div>
							<div className="text-center">
								<h3 className="font-semibold text-lg">
									2FA Enabled!
								</h3>
								<p className="text-sm text-muted-foreground mt-1">
									Your account is now protected with
									two-factor authentication.
								</p>
							</div>
						</div>
					)}

					{/* Disable 2FA Step */}
					{setupStep === "disabling" && (
						<form onSubmit={handleDisable} className="space-y-4">
							<div className="p-3 bg-destructive/10 border border-destructive/20 rounded-lg">
								<p className="text-sm text-destructive">
									<strong>Warning:</strong> Disabling
									two-factor authentication will make your
									account less secure.
								</p>
							</div>

							<div className="space-y-2">
								<Label htmlFor="disable-password">
									Confirm your password to disable 2FA
								</Label>
								<Input
									id="disable-password"
									name="password"
									type="password"
									placeholder="Enter your password"
									value={password}
									onChange={(e) => {
										setPassword(e.target.value);
										setError(null);
									}}
									disabled={isLoading}
									required
									autoComplete="current-password"
									className="h-11"
								/>
							</div>

							<div className="flex gap-3">
								<Button
									type="button"
									variant="outline"
									onClick={resetState}
									disabled={isLoading}
								>
									Cancel
								</Button>
								<Button
									type="submit"
									variant="destructive"
									disabled={isLoading || !password}
								>
									{isLoading ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Disabling...
										</>
									) : (
										<>
											<ShieldOff className="mr-2 h-4 w-4" />
											Disable 2FA
										</>
									)}
								</Button>
							</div>
						</form>
					)}
				</CardContent>
			</Card>

			{/* Password Change Section */}
			<PasswordChangeSection />
		</div>
	);
}

function PasswordChangeSection() {
	const [isChangingPassword, setIsChangingPassword] = useState(false);
	const [currentPassword, setCurrentPassword] = useState("");
	const [newPassword, setNewPassword] = useState("");
	const [confirmPassword, setConfirmPassword] = useState("");
	const [passwordError, setPasswordError] = useState<string | null>(null);
	const [passwordSuccess, setPasswordSuccess] = useState(false);
	const [isPasswordLoading, setIsPasswordLoading] = useState(false);

	const resetPasswordForm = () => {
		setIsChangingPassword(false);
		setCurrentPassword("");
		setNewPassword("");
		setConfirmPassword("");
		setPasswordError(null);
		setPasswordSuccess(false);
	};

	const handleChangePassword = async (
		e: React.FormEvent<HTMLFormElement>,
	) => {
		e.preventDefault();
		setPasswordError(null);

		if (newPassword !== confirmPassword) {
			setPasswordError("New passwords do not match");
			return;
		}

		if (newPassword.length < 8) {
			setPasswordError("New password must be at least 8 characters");
			return;
		}

		setIsPasswordLoading(true);

		try {
			const response = await fetch("/api/auth/change-password", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				credentials: "include",
				body: JSON.stringify({
					currentPassword,
					newPassword,
				}),
			});

			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw new Error(data.message || "Failed to change password");
			}

			setPasswordSuccess(true);
			setTimeout(() => {
				resetPasswordForm();
			}, 2000);
		} catch (err) {
			setPasswordError(
				err instanceof Error
					? err.message
					: "An error occurred. Please try again.",
			);
		} finally {
			setIsPasswordLoading(false);
		}
	};

	if (passwordSuccess) {
		return (
			<Card>
				<CardHeader>
					<CardTitle>Password</CardTitle>
					<CardDescription>
						Change your password to keep your account secure
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex items-center gap-2 p-3 bg-primary/10 rounded-lg text-primary">
						<CheckCircle2 className="h-5 w-5" />
						<span className="text-sm font-medium">
							Password changed successfully!
						</span>
					</div>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle>Password</CardTitle>
				<CardDescription>
					Change your password to keep your account secure
				</CardDescription>
			</CardHeader>
			<CardContent>
				{!isChangingPassword ? (
					<Button
						variant="outline"
						onClick={() => setIsChangingPassword(true)}
					>
						<KeyRound className="mr-2 h-4 w-4" />
						Change password
					</Button>
				) : (
					<form onSubmit={handleChangePassword} className="space-y-4">
						{passwordError && (
							<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive flex items-start gap-2">
								<AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
								{passwordError}
							</div>
						)}

						<div className="space-y-2">
							<Label htmlFor="currentPassword">
								Current password
							</Label>
							<Input
								id="currentPassword"
								type="password"
								placeholder="Enter current password"
								value={currentPassword}
								onChange={(e) => {
									setCurrentPassword(e.target.value);
									setPasswordError(null);
								}}
								disabled={isPasswordLoading}
								required
								autoComplete="current-password"
								className="h-11"
							/>
						</div>

						<div className="space-y-2">
							<Label htmlFor="newPassword">New password</Label>
							<Input
								id="newPassword"
								type="password"
								placeholder="Enter new password"
								value={newPassword}
								onChange={(e) => {
									setNewPassword(e.target.value);
									setPasswordError(null);
								}}
								disabled={isPasswordLoading}
								required
								minLength={8}
								autoComplete="new-password"
								className="h-11"
							/>
						</div>

						<div className="space-y-2">
							<Label htmlFor="confirmPassword">
								Confirm new password
							</Label>
							<Input
								id="confirmPassword"
								type="password"
								placeholder="Confirm new password"
								value={confirmPassword}
								onChange={(e) => {
									setConfirmPassword(e.target.value);
									setPasswordError(null);
								}}
								disabled={isPasswordLoading}
								required
								minLength={8}
								autoComplete="new-password"
								className="h-11"
							/>
						</div>

						<div className="flex gap-3">
							<Button
								type="button"
								variant="outline"
								onClick={resetPasswordForm}
								disabled={isPasswordLoading}
							>
								Cancel
							</Button>
							<Button
								type="submit"
								disabled={
									isPasswordLoading ||
									!currentPassword ||
									!newPassword ||
									!confirmPassword
								}
							>
								{isPasswordLoading ? (
									<>
										<Loader2 className="mr-2 h-4 w-4 animate-spin" />
										Changing...
									</>
								) : (
									"Change password"
								)}
							</Button>
						</div>
					</form>
				)}
			</CardContent>
		</Card>
	);
}
