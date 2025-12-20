"use client";

import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { signUp, signIn } from "@/lib/auth-client";
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
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import {
	Loader2,
	Github,
	UserPlus,
	CheckCircle2,
	AlertTriangle,
} from "lucide-react";

function RegisterForm() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const invitedEmail = searchParams.get("email");
	const isInvited = searchParams.get("invited") === "true";

	const [isLoading, setIsLoading] = useState(false);
	const [isGithubLoading, setIsGithubLoading] = useState(false);
	const [isGoogleLoading, setIsGoogleLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [success, setSuccess] = useState<string | null>(null);

	const [formData, setFormData] = useState({
		name: "",
		email: invitedEmail || "",
		password: "",
		confirmPassword: "",
	});

	// Lock email field if invited
	const isEmailLocked = isInvited && !!invitedEmail;

	useEffect(() => {
		if (invitedEmail) {
			setFormData((prev) => ({ ...prev, email: invitedEmail }));
		}
	}, [invitedEmail]);

	const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		if (name === "email" && isEmailLocked) return;
		setFormData((prev) => ({ ...prev, [name]: value }));
		setError(null);
	};

	const validateForm = () => {
		if (!formData.name.trim()) {
			setError("Please enter your name");
			return false;
		}
		if (!formData.email.trim()) {
			setError("Please enter your email");
			return false;
		}
		if (formData.password.length < 8) {
			setError("Password must be at least 8 characters");
			return false;
		}
		if (formData.password !== formData.confirmPassword) {
			setError("Passwords do not match");
			return false;
		}
		return true;
	};

	const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		if (!validateForm()) return;

		setIsLoading(true);
		setError(null);

		try {
			const result = await signUp.email({
				email: formData.email,
				password: formData.password,
				name: formData.name,
			});

			if (result.error) {
				// Check for unauthorized error
				if (
					result.error.message?.includes("UNAUTHORIZED") ||
					result.error.message?.includes("not authorized")
				) {
					setError(
						"This email is not authorized. Please request an invite from an administrator.",
					);
				} else {
					setError(
						result.error.message ||
							"Error creating account. Please try again.",
					);
				}
				return;
			}

			setSuccess("Account created successfully! Redirecting...");

			// Auto login after registration
			setTimeout(() => {
				router.push("/dashboard");
				router.refresh();
			}, 1500);
		} catch {
			setError("An error occurred. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleGithubSignUp = async () => {
		if (!isInvited) {
			setError("Only invited users can register.");
			return;
		}
		setIsGithubLoading(true);
		setError(null);

		try {
			await signIn.social({
				provider: "github",
				callbackURL: "/dashboard",
			});
		} catch {
			setError("Error connecting to GitHub.");
			setIsGithubLoading(false);
		}
	};

	const handleGoogleSignUp = async () => {
		if (!isInvited) {
			setError("Only invited users can register.");
			return;
		}
		setIsGoogleLoading(true);
		setError(null);

		try {
			await signIn.social({
				provider: "google",
				callbackURL: "/dashboard",
			});
		} catch {
			setError("Error connecting to Google.");
			setIsGoogleLoading(false);
		}
	};

	const isAnyLoading = isLoading || isGithubLoading || isGoogleLoading;

	// If not invited and no email, show unauthorized message
	if (!isInvited && !invitedEmail) {
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
							Restricted Access
						</CardTitle>
						<CardDescription className="text-muted-foreground">
							This platform is invite-only
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-6">
						<div className="rounded-lg bg-amber-500/10 border border-amber-500/20 p-4 text-sm text-amber-600 dark:text-amber-400 flex items-start gap-3">
							<AlertTriangle className="h-5 w-5 mt-0.5 shrink-0" />
							<div>
								<p className="font-medium">
									Invitation required
								</p>
								<p className="mt-1 text-amber-600/80 dark:text-amber-400/80">
									To access the platform, you need an
									invitation from an administrator. Contact
									the administrator to request access.
								</p>
							</div>
						</div>

						<div className="text-center">
							<p className="text-sm text-muted-foreground mb-4">
								Already have an invite or an account?
							</p>
							<Button variant="outline" asChild>
								<Link href="/login">Sign in</Link>
							</Button>
						</div>
					</CardContent>
				</Card>
			</div>
		);
	}

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
				<CardHeader className="space-y-1 pb-6">
					<CardTitle className="text-2xl font-bold tracking-tight">
						Create your account
					</CardTitle>
					<CardDescription className="text-muted-foreground">
						{isInvited
							? "Complete your registration to access the platform"
							: "Fill in the details below to create your account"}
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-6">
					{isInvited && (
						<div className="rounded-lg bg-primary/10 border border-primary/20 p-3 text-sm text-primary flex items-start gap-2">
							<CheckCircle2 className="h-4 w-4 mt-0.5 shrink-0" />
							<span>
								You have been invited to access the platform.
								Complete your registration below.
							</span>
						</div>
					)}

					{/* Social Login Buttons */}
					<div className="grid grid-cols-2 gap-3">
						<Button
							variant="outline"
							type="button"
							disabled={isAnyLoading}
							onClick={handleGithubSignUp}
							className="h-11"
						>
							{isGithubLoading ? (
								<Loader2 className="h-4 w-4 animate-spin" />
							) : (
								<Github className="h-4 w-4" />
							)}
							<span className="ml-2">GitHub</span>
						</Button>
						<Button
							variant="outline"
							type="button"
							disabled={isAnyLoading}
							onClick={handleGoogleSignUp}
							className="h-11"
						>
							{isGoogleLoading ? (
								<Loader2 className="h-4 w-4 animate-spin" />
							) : (
								<svg
									className="h-4 w-4"
									viewBox="0 0 24 24"
									aria-hidden="true"
								>
									<path
										d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
										fill="#4285F4"
									/>
									<path
										d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
										fill="#34A853"
									/>
									<path
										d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
										fill="#FBBC05"
									/>
									<path
										d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
										fill="#EA4335"
									/>
								</svg>
							)}
							<span className="ml-2">Google</span>
						</Button>
					</div>

					<div className="relative">
						<div className="absolute inset-0 flex items-center">
							<Separator className="w-full" />
						</div>
						<div className="relative flex justify-center text-xs uppercase">
							<span className="bg-card px-2 text-muted-foreground">
								or register with email
							</span>
						</div>
					</div>

					{/* Registration Form */}
					<form onSubmit={handleSubmit} className="space-y-4">
						{error && (
							<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive">
								{error}
							</div>
						)}

						{success && (
							<div className="rounded-lg bg-primary/10 border border-primary/20 p-3 text-sm text-primary flex items-center gap-2">
								<CheckCircle2 className="h-4 w-4" />
								{success}
							</div>
						)}

						<div className="space-y-2">
							<Label htmlFor="name">Name</Label>
							<Input
								id="name"
								name="name"
								type="text"
								placeholder="Your full name"
								value={formData.name}
								onChange={handleInputChange}
								disabled={isAnyLoading}
								required
								autoComplete="name"
								className="h-11"
							/>
						</div>

						<div className="space-y-2">
							<Label htmlFor="email">Email</Label>
							<Input
								id="email"
								name="email"
								type="email"
								placeholder="your@email.com"
								value={formData.email}
								onChange={handleInputChange}
								disabled={isAnyLoading || isEmailLocked}
								required
								autoComplete="email"
								className="h-11"
							/>
							{isEmailLocked && (
								<p className="text-xs text-muted-foreground">
									This email was set by the invitation and
									cannot be changed.
								</p>
							)}
						</div>

						<div className="space-y-2">
							<Label htmlFor="password">Password</Label>
							<Input
								id="password"
								name="password"
								type="password"
								placeholder="Minimum 8 characters"
								value={formData.password}
								onChange={handleInputChange}
								disabled={isAnyLoading}
								required
								autoComplete="new-password"
								className="h-11"
							/>
						</div>

						<div className="space-y-2">
							<Label htmlFor="confirmPassword">
								Confirm Password
							</Label>
							<Input
								id="confirmPassword"
								name="confirmPassword"
								type="password"
								placeholder="Repeat your password"
								value={formData.confirmPassword}
								onChange={handleInputChange}
								disabled={isAnyLoading}
								required
								autoComplete="new-password"
								className="h-11"
							/>
						</div>

						<Button
							type="submit"
							className="w-full h-11 font-medium"
							disabled={isAnyLoading}
						>
							{isLoading ? (
								<>
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
									Creating account...
								</>
							) : (
								<>
									<UserPlus className="mr-2 h-4 w-4" />
									Create account
								</>
							)}
						</Button>
					</form>

					<div className="text-center text-sm text-muted-foreground">
						Already have an account?{" "}
						<Link
							href="/login"
							className="text-primary hover:text-primary/80 transition-colors font-medium"
						>
							Sign in
						</Link>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}

function RegisterSkeleton() {
	return (
		<div className="space-y-6">
			<Card className="border-0 shadow-none lg:border lg:shadow-sm">
				<CardHeader className="space-y-1 pb-6">
					<Skeleton className="h-8 w-48" />
					<Skeleton className="h-4 w-72" />
				</CardHeader>
				<CardContent className="space-y-6">
					<div className="grid grid-cols-2 gap-3">
						<Skeleton className="h-11" />
						<Skeleton className="h-11" />
					</div>
					<Skeleton className="h-px w-full" />
					<div className="space-y-4">
						<Skeleton className="h-11" />
						<Skeleton className="h-11" />
						<Skeleton className="h-11" />
						<Skeleton className="h-11" />
						<Skeleton className="h-11" />
					</div>
				</CardContent>
			</Card>
		</div>
	);
}

export default function RegisterPage() {
	return (
		<Suspense fallback={<RegisterSkeleton />}>
			<RegisterForm />
		</Suspense>
	);
}
