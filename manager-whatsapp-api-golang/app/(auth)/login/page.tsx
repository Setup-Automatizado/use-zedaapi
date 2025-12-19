"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { signIn } from "@/lib/auth-client";
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
import { Loader2, Github, Mail } from "lucide-react";

export default function LoginPage() {
	const router = useRouter();
	const [isLoading, setIsLoading] = useState(false);
	const [isGithubLoading, setIsGithubLoading] = useState(false);
	const [isGoogleLoading, setIsGoogleLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const [formData, setFormData] = useState({
		email: "",
		password: "",
	});

	const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		setFormData((prev) => ({ ...prev, [name]: value }));
		setError(null);
	};

	const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		setIsLoading(true);
		setError(null);

		try {
			const result = await signIn.email({
				email: formData.email,
				password: formData.password,
			});

			if (result.error) {
				setError(
					result.error.message ||
						"Invalid credentials. Please try again.",
				);
				return;
			}

			router.push("/dashboard");
			router.refresh();
		} catch {
			setError("An error occurred while logging in. Please try again.");
		} finally {
			setIsLoading(false);
		}
	};

	const handleGithubSignIn = async () => {
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

	const handleGoogleSignIn = async () => {
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
						Welcome back
					</CardTitle>
					<CardDescription className="text-muted-foreground">
						Enter your credentials to access the dashboard
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-6">
					{/* Social Login Buttons */}
					<div className="grid grid-cols-2 gap-3">
						<Button
							variant="outline"
							type="button"
							disabled={isAnyLoading}
							onClick={handleGithubSignIn}
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
							onClick={handleGoogleSignIn}
							className="h-11"
						>
							{isGoogleLoading ? (
								<Loader2 className="h-4 w-4 animate-spin" />
							) : (
								<svg className="h-4 w-4" viewBox="0 0 24 24">
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
								or continue with email
							</span>
						</div>
					</div>

					{/* Email/Password Form */}
					<form onSubmit={handleSubmit} className="space-y-4">
						{error && (
							<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive">
								{error}
							</div>
						)}

						<div className="space-y-2">
							<Label htmlFor="email">Email</Label>
							<Input
								id="email"
								name="email"
								type="email"
								placeholder="your@email.com"
								value={formData.email}
								onChange={handleInputChange}
								disabled={isAnyLoading}
								required
								autoComplete="email"
								className="h-11"
							/>
						</div>

						<div className="space-y-2">
							<div className="flex items-center justify-between">
								<Label htmlFor="password">Password</Label>
								<Link
									href="/forgot-password"
									className="text-sm text-primary hover:text-primary/80 transition-colors"
								>
									Forgot password?
								</Link>
							</div>
							<Input
								id="password"
								name="password"
								type="password"
								placeholder="Your password"
								value={formData.password}
								onChange={handleInputChange}
								disabled={isAnyLoading}
								required
								autoComplete="current-password"
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
									Signing in...
								</>
							) : (
								<>
									<Mail className="mr-2 h-4 w-4" />
									Sign in with Email
								</>
							)}
						</Button>
					</form>
				</CardContent>
			</Card>
		</div>
	);
}
