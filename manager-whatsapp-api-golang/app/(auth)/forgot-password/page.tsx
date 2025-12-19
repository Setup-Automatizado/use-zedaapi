"use client";

import { useState } from "react";
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
import { Loader2, ArrowLeft, Mail, CheckCircle2 } from "lucide-react";

export default function ForgotPasswordPage() {
	const [isLoading, setIsLoading] = useState(false);
	const [isSuccess, setIsSuccess] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [email, setEmail] = useState("");

	const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		setIsLoading(true);
		setError(null);

		try {
			const response = await fetch("/api/auth/request-password-reset", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({
					email,
					redirectTo: "/reset-password",
				}),
			});

			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw new Error(data.message || "Error sending recovery email");
			}

			setIsSuccess(true);
		} catch (err) {
			setError(
				err instanceof Error
					? err.message
					: "An error occurred while sending the email. Please try again.",
			);
		} finally {
			setIsLoading(false);
		}
	};

	if (isSuccess) {
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
							<CheckCircle2 className="h-8 w-8 text-primary" />
						</div>
						<CardTitle className="text-2xl font-bold tracking-tight">
							Email sent!
						</CardTitle>
						<CardDescription className="text-muted-foreground">
							We sent a recovery link to <strong>{email}</strong>.
							Check your inbox.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						<p className="text-sm text-muted-foreground text-center">
							Didn&apos;t receive the email? Check your spam folder or
							try again.
						</p>
						<div className="flex flex-col gap-3">
							<Button
								variant="outline"
								onClick={() => setIsSuccess(false)}
								className="h-11"
							>
								Try again
							</Button>
							<Link href="/login" className="w-full">
								<Button variant="ghost" className="w-full h-11">
									<ArrowLeft className="mr-2 h-4 w-4" />
									Back to login
								</Button>
							</Link>
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
						Forgot your password?
					</CardTitle>
					<CardDescription className="text-muted-foreground">
						Enter your email and we&apos;ll send you a link to reset your
						password.
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
							<Label htmlFor="email">Email</Label>
							<Input
								id="email"
								name="email"
								type="email"
								placeholder="your@email.com"
								value={email}
								onChange={(e) => {
									setEmail(e.target.value);
									setError(null);
								}}
								disabled={isLoading}
								required
								autoComplete="email"
								className="h-11"
							/>
						</div>

						<Button
							type="submit"
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
									Send recovery link
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
