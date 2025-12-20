/**
 * Profile Settings Page
 *
 * Allows users to update their profile information,
 * upload avatar, and manage linked social accounts.
 */

"use client";

import * as React from "react";
import { useSession, signIn } from "@/lib/auth-client";
import { useRouter } from "next/navigation";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import {
	AlertCircle,
	Camera,
	Check,
	Github,
	Loader2,
	Trash2,
	Link as LinkIcon,
	Unlink,
	User,
	Calendar,
	Shield,
} from "lucide-react";

// Google icon component
function GoogleIcon({ className }: { className?: string }) {
	return (
		<svg className={className} viewBox="0 0 24 24" aria-hidden="true">
			<path
				fill="currentColor"
				d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
			/>
			<path
				fill="currentColor"
				d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
			/>
			<path
				fill="currentColor"
				d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
			/>
			<path
				fill="currentColor"
				d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
			/>
		</svg>
	);
}

interface LinkedAccount {
	id: string;
	provider: string;
	linkedAt: string;
}

interface ProfileData {
	id: string;
	name: string | null;
	email: string;
	image: string | null;
	role: string;
	createdAt: string;
	linkedAccounts: LinkedAccount[];
}

export default function ProfileSettingsPage() {
	const { data: session, isPending: sessionLoading } = useSession();
	const router = useRouter();
	const fileInputRef = React.useRef<HTMLInputElement>(null);

	// Form state
	const [name, setName] = React.useState("");
	const [profile, setProfile] = React.useState<ProfileData | null>(null);

	// Loading states
	const [isLoading, setIsLoading] = React.useState(true);
	const [isSaving, setIsSaving] = React.useState(false);
	const [isUploading, setIsUploading] = React.useState(false);
	const [isUnlinking, setIsUnlinking] = React.useState<string | null>(null);

	// Messages
	const [error, setError] = React.useState<string | null>(null);
	const [success, setSuccess] = React.useState<string | null>(null);

	// Fetch profile data
	const fetchProfile = React.useCallback(async () => {
		try {
			const response = await fetch("/api/profile");
			if (!response.ok) throw new Error("Failed to fetch profile");
			const data = await response.json();
			setProfile(data);
			setName(data.name || "");
		} catch {
			setError("Failed to load profile");
		} finally {
			setIsLoading(false);
		}
	}, []);

	React.useEffect(() => {
		if (!sessionLoading && !session) {
			router.push("/login");
			return;
		}

		if (session) {
			fetchProfile();
		}
	}, [session, sessionLoading, router, fetchProfile]);

	// Handle name update
	const handleSaveName = async () => {
		if (!name.trim()) {
			setError("Name is required");
			return;
		}

		setIsSaving(true);
		setError(null);
		setSuccess(null);

		try {
			const response = await fetch("/api/profile", {
				method: "PATCH",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ name: name.trim() }),
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || "Failed to update profile");
			}

			setSuccess("Profile updated successfully");
			fetchProfile();
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Failed to update profile",
			);
		} finally {
			setIsSaving(false);
		}
	};

	// Handle avatar upload
	const handleAvatarUpload = async (
		e: React.ChangeEvent<HTMLInputElement>,
	) => {
		const file = e.target.files?.[0];
		if (!file) return;

		setIsUploading(true);
		setError(null);
		setSuccess(null);

		try {
			const formData = new FormData();
			formData.append("file", file);

			const response = await fetch("/api/upload/avatar", {
				method: "POST",
				body: formData,
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || "Failed to upload avatar");
			}

			setSuccess("Avatar updated successfully");
			fetchProfile();
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Failed to upload avatar",
			);
		} finally {
			setIsUploading(false);
			if (fileInputRef.current) {
				fileInputRef.current.value = "";
			}
		}
	};

	// Handle avatar removal
	const handleRemoveAvatar = async () => {
		setIsUploading(true);
		setError(null);
		setSuccess(null);

		try {
			const response = await fetch("/api/upload/avatar", {
				method: "DELETE",
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || "Failed to remove avatar");
			}

			setSuccess("Avatar removed successfully");
			fetchProfile();
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Failed to remove avatar",
			);
		} finally {
			setIsUploading(false);
		}
	};

	// Handle social account linking
	const handleLinkAccount = async (provider: "github" | "google") => {
		try {
			await signIn.social({
				provider,
				callbackURL: "/settings/profile",
			});
		} catch {
			setError(`Failed to link ${provider} account`);
		}
	};

	// Handle social account unlinking
	const handleUnlinkAccount = async (accountId: string) => {
		setIsUnlinking(accountId);
		setError(null);
		setSuccess(null);

		try {
			const response = await fetch(`/api/profile/accounts/${accountId}`, {
				method: "DELETE",
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || "Failed to unlink account");
			}

			setSuccess("Account unlinked successfully");
			fetchProfile();
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Failed to unlink account",
			);
		} finally {
			setIsUnlinking(null);
		}
	};

	// Get initials for avatar fallback
	const getInitials = (name: string | null, email: string) => {
		if (name) {
			return name
				.split(" ")
				.map((n) => n[0])
				.join("")
				.toUpperCase()
				.slice(0, 2);
		}
		return email.slice(0, 2).toUpperCase();
	};

	// Check if provider is linked
	const isProviderLinked = (provider: string) => {
		return profile?.linkedAccounts.some((a) => a.provider === provider);
	};

	// Get linked account by provider
	const getLinkedAccount = (provider: string) => {
		return profile?.linkedAccounts.find((a) => a.provider === provider);
	};

	if (sessionLoading || isLoading) {
		return (
			<div className="flex items-center justify-center min-h-[400px]">
				<Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
			</div>
		);
	}

	return (
		<div className="space-y-6">
			{/* Header */}
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Profile</h1>
				<p className="text-muted-foreground">
					Manage your account settings
				</p>
			</div>

			{/* Messages */}
			{error && (
				<Alert variant="destructive">
					<AlertCircle className="h-4 w-4" />
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			)}

			{success && (
				<Alert className="border-green-500 bg-green-500/10">
					<Check className="h-4 w-4 text-green-500" />
					<AlertDescription className="text-green-500">
						{success}
					</AlertDescription>
				</Alert>
			)}

			<div className="grid gap-6 lg:grid-cols-3">
				{/* Left Column - Avatar & Info */}
				<Card className="lg:col-span-1">
					<CardContent className="pt-6">
						<div className="flex flex-col items-center text-center">
							{/* Avatar */}
							<div className="relative group">
								<Avatar className="h-24 w-24">
									<AvatarImage
										src={profile?.image || undefined}
										alt={profile?.name || "Avatar"}
									/>
									<AvatarFallback className="text-xl bg-primary text-primary-foreground">
										{getInitials(
											profile?.name || null,
											profile?.email || "",
										)}
									</AvatarFallback>
								</Avatar>
								{isUploading && (
									<div className="absolute inset-0 flex items-center justify-center bg-black/50 rounded-full">
										<Loader2 className="h-6 w-6 animate-spin text-white" />
									</div>
								)}
								<input
									ref={fileInputRef}
									type="file"
									accept="image/jpeg,image/png,image/gif,image/webp"
									onChange={handleAvatarUpload}
									className="hidden"
								/>
								<button
									onClick={() =>
										fileInputRef.current?.click()
									}
									disabled={isUploading}
									className="absolute bottom-0 right-0 p-1.5 rounded-full bg-primary text-primary-foreground shadow-lg hover:bg-primary/90 transition-colors"
								>
									<Camera className="h-3.5 w-3.5" />
								</button>
							</div>

							{/* Name & Email */}
							<h3 className="mt-4 font-semibold text-lg">
								{profile?.name || "No name set"}
							</h3>
							<p className="text-sm text-muted-foreground">
								{profile?.email}
							</p>

							{/* Role Badge */}
							<Badge variant="secondary" className="mt-2">
								<Shield className="h-3 w-3 mr-1" />
								{profile?.role || "User"}
							</Badge>

							{/* Remove Avatar Button */}
							{profile?.image && (
								<Button
									variant="ghost"
									size="sm"
									onClick={handleRemoveAvatar}
									disabled={isUploading}
									className="mt-3 text-muted-foreground hover:text-destructive"
								>
									<Trash2 className="h-3.5 w-3.5 mr-1.5" />
									Remove photo
								</Button>
							)}

							<Separator className="my-4 w-full" />

							{/* Account Info */}
							<div className="w-full space-y-3 text-sm">
								<div className="flex items-center justify-between">
									<span className="text-muted-foreground flex items-center gap-2">
										<Calendar className="h-4 w-4" />
										Member since
									</span>
									<span className="font-medium">
										{profile?.createdAt
											? new Date(
													profile.createdAt,
												).toLocaleDateString("en-US", {
													month: "short",
													year: "numeric",
												})
											: "-"}
									</span>
								</div>
							</div>
						</div>
					</CardContent>
				</Card>

				{/* Right Column - Forms */}
				<div className="lg:col-span-2 space-y-6">
					{/* Personal Information */}
					<Card>
						<CardHeader className="pb-4">
							<CardTitle className="text-base flex items-center gap-2">
								<User className="h-4 w-4" />
								Personal Information
							</CardTitle>
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="grid gap-4 sm:grid-cols-2">
								<div className="space-y-2">
									<Label htmlFor="name">Display Name</Label>
									<Input
										id="name"
										value={name}
										onChange={(e) =>
											setName(e.target.value)
										}
										placeholder="Enter your name"
										disabled={isSaving}
									/>
								</div>
								<div className="space-y-2">
									<Label htmlFor="email">Email</Label>
									<Input
										id="email"
										value={profile?.email || ""}
										disabled
										className="bg-muted"
									/>
								</div>
							</div>
							<div className="flex justify-end">
								<Button
									onClick={handleSaveName}
									disabled={isSaving}
									size="sm"
								>
									{isSaving ? (
										<>
											<Loader2 className="h-4 w-4 mr-2 animate-spin" />
											Saving...
										</>
									) : (
										"Save Changes"
									)}
								</Button>
							</div>
						</CardContent>
					</Card>

					{/* Linked Accounts */}
					<Card>
						<CardHeader className="pb-4">
							<CardTitle className="text-base flex items-center gap-2">
								<LinkIcon className="h-4 w-4" />
								Linked Accounts
							</CardTitle>
							<CardDescription>
								Connect social accounts for easier login
							</CardDescription>
						</CardHeader>
						<CardContent>
							<div className="space-y-3">
								{/* GitHub */}
								<div className="flex items-center justify-between p-3 rounded-lg border bg-muted/30">
									<div className="flex items-center gap-3">
										<div className="flex h-9 w-9 items-center justify-center rounded-full bg-background border">
											<Github className="h-4 w-4" />
										</div>
										<div>
											<p className="font-medium text-sm">
												GitHub
											</p>
											<p className="text-xs text-muted-foreground">
												{isProviderLinked("github")
													? "Connected"
													: "Not connected"}
											</p>
										</div>
									</div>
									{isProviderLinked("github") ? (
										<Button
											variant="outline"
											size="sm"
											onClick={() =>
												handleUnlinkAccount(
													getLinkedAccount("github")!
														.id,
												)
											}
											disabled={
												isUnlinking ===
												getLinkedAccount("github")?.id
											}
										>
											{isUnlinking ===
											getLinkedAccount("github")?.id ? (
												<Loader2 className="h-4 w-4 animate-spin" />
											) : (
												<>
													<Unlink className="h-3.5 w-3.5 mr-1.5" />
													Unlink
												</>
											)}
										</Button>
									) : (
										<Button
											variant="outline"
											size="sm"
											onClick={() =>
												handleLinkAccount("github")
											}
										>
											<LinkIcon className="h-3.5 w-3.5 mr-1.5" />
											Link
										</Button>
									)}
								</div>

								{/* Google */}
								<div className="flex items-center justify-between p-3 rounded-lg border bg-muted/30">
									<div className="flex items-center gap-3">
										<div className="flex h-9 w-9 items-center justify-center rounded-full bg-background border">
											<GoogleIcon className="h-4 w-4" />
										</div>
										<div>
											<p className="font-medium text-sm">
												Google
											</p>
											<p className="text-xs text-muted-foreground">
												{isProviderLinked("google")
													? "Connected"
													: "Not connected"}
											</p>
										</div>
									</div>
									{isProviderLinked("google") ? (
										<Button
											variant="outline"
											size="sm"
											onClick={() =>
												handleUnlinkAccount(
													getLinkedAccount("google")!
														.id,
												)
											}
											disabled={
												isUnlinking ===
												getLinkedAccount("google")?.id
											}
										>
											{isUnlinking ===
											getLinkedAccount("google")?.id ? (
												<Loader2 className="h-4 w-4 animate-spin" />
											) : (
												<>
													<Unlink className="h-3.5 w-3.5 mr-1.5" />
													Unlink
												</>
											)}
										</Button>
									) : (
										<Button
											variant="outline"
											size="sm"
											onClick={() =>
												handleLinkAccount("google")
											}
										>
											<LinkIcon className="h-3.5 w-3.5 mr-1.5" />
											Link
										</Button>
									)}
								</div>
							</div>
						</CardContent>
					</Card>
				</div>
			</div>
		</div>
	);
}
