"use client";

import { useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
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
import {
	Loader2,
	ArrowLeft,
	Lock,
	CheckCircle2,
	Eye,
	EyeOff,
} from "lucide-react";
import authClient from "@/lib/auth-client";

function ResetPasswordForm() {
	const searchParams = useSearchParams();
	const router = useRouter();
	const token = searchParams.get("token");

	const [isLoading, setIsLoading] = useState(false);
	const [isSuccess, setIsSuccess] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [password, setPassword] = useState("");
	const [confirmPassword, setConfirmPassword] = useState("");
	const [showPassword, setShowPassword] = useState(false);
	const [showConfirmPassword, setShowConfirmPassword] = useState(false);

	const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		setIsLoading(true);
		setError(null);

		if (password !== confirmPassword) {
			setError("Passwords do not match");
			setIsLoading(false);
			return;
		}

		if (password.length < 8) {
			setError("Password must be at least 8 characters");
			setIsLoading(false);
			return;
		}

		if (!token) {
			setError("Invalid or missing recovery token");
			setIsLoading(false);
			return;
		}

		try {
			const { error: resetError } = await authClient.resetPassword({
				newPassword: password,
				token,
			});

			if (resetError) {
				throw new Error(
					resetError.message || "Error resetting password",
				);
			}

			setIsSuccess(true);
		} catch (err) {
			setError(
				err instanceof Error
					? err.message
					: "An error occurred while resetting your password. Please try again.",
			);
		} finally {
			setIsLoading(false);
		}
	};

	// Invalid token
	if (!token) {
		return (
			<div className="space-y-6">
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
						<CardTitle className="text-2xl font-bold tracking-tight">
							Invalid link
						</CardTitle>
						<CardDescription className="text-muted-foreground">
							The recovery link is invalid or has expired. Please
							request a new one.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Link href="/forgot-password" className="block">
							<Button className="w-full h-11">
								Request new link
							</Button>
						</Link>
					</CardContent>
				</Card>
			</div>
		);
	}

	// Success
	if (isSuccess) {
		return (
			<div className="space-y-6">
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
							<CheckCircle2 className="h-8 w-8 text-primary" />
						</div>
						<CardTitle className="text-2xl font-bold tracking-tight">
							Password reset!
						</CardTitle>
						<CardDescription className="text-muted-foreground">
							Your password has been changed successfully. You can
							now login.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Button
							onClick={() => router.push("/login")}
							className="w-full h-11"
						>
							Go to login
						</Button>
					</CardContent>
				</Card>
			</div>
		);
	}

	// Form
	return (
		<div className="space-y-6">
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
				<CardHeader className="space-y-1 pb-6">
					<CardTitle className="text-2xl font-bold tracking-tight">
						Reset password
					</CardTitle>
					<CardDescription className="text-muted-foreground">
						Enter your new password below.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-6">
					<form onSubmit={handleSubmit} className="space-y-4">
						{error && (
							<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive">
								{error}
							</div>
						)}

						<div className="space-y-2">
							<Label htmlFor="password">New password</Label>
							<div className="relative">
								<Input
									id="password"
									name="password"
									type={showPassword ? "text" : "password"}
									placeholder="********"
									value={password}
									onChange={(e) => {
										setPassword(e.target.value);
										setError(null);
									}}
									disabled={isLoading}
									required
									minLength={8}
									autoComplete="new-password"
									className="h-11 pr-10"
								/>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									className="absolute right-0 top-0 h-11 px-3 hover:bg-transparent"
									onClick={() =>
										setShowPassword(!showPassword)
									}
								>
									{showPassword ? (
										<EyeOff className="h-4 w-4 text-muted-foreground" />
									) : (
										<Eye className="h-4 w-4 text-muted-foreground" />
									)}
								</Button>
							</div>
						</div>

						<div className="space-y-2">
							<Label htmlFor="confirmPassword">
								Confirm new password
							</Label>
							<div className="relative">
								<Input
									id="confirmPassword"
									name="confirmPassword"
									type={
										showConfirmPassword
											? "text"
											: "password"
									}
									placeholder="********"
									value={confirmPassword}
									onChange={(e) => {
										setConfirmPassword(e.target.value);
										setError(null);
									}}
									disabled={isLoading}
									required
									minLength={8}
									autoComplete="new-password"
									className="h-11 pr-10"
								/>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									className="absolute right-0 top-0 h-11 px-3 hover:bg-transparent"
									onClick={() =>
										setShowConfirmPassword(
											!showConfirmPassword,
										)
									}
								>
									{showConfirmPassword ? (
										<EyeOff className="h-4 w-4 text-muted-foreground" />
									) : (
										<Eye className="h-4 w-4 text-muted-foreground" />
									)}
								</Button>
							</div>
						</div>

						<Button
							type="submit"
							className="w-full h-11 font-medium"
							disabled={isLoading}
						>
							{isLoading ? (
								<>
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
									Resetting...
								</>
							) : (
								<>
									<Lock className="mr-2 h-4 w-4" />
									Reset password
								</>
							)}
						</Button>
					</form>

					<Link href="/login" className="block">
						<Button variant="ghost" className="w-full h-11">
							<ArrowLeft className="mr-2 h-4 w-4" />
							Back to login
						</Button>
					</Link>
				</CardContent>
			</Card>
		</div>
	);
}

export default function ResetPasswordPage() {
	return (
		<Suspense
			fallback={
				<div className="flex items-center justify-center min-h-[400px]">
					<Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
				</div>
			}
		>
			<ResetPasswordForm />
		</Suspense>
	);
}
