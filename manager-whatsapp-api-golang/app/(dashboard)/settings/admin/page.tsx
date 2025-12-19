"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/lib/auth-client";
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
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
	Loader2,
	UserPlus,
	Trash2,
	Shield,
	Users,
	Mail,
	CheckCircle2,
	Clock,
	AlertTriangle,
} from "lucide-react";

interface Invitation {
	id: string;
	email: string;
	role: string;
	invitedAt: string;
	acceptedAt: string | null;
	invitedBy: {
		name: string | null;
		email: string;
	} | null;
	user: {
		name: string | null;
		email: string;
		createdAt: string;
	} | null;
}

export default function AdminSettingsPage() {
	const { data: session, isPending: isSessionLoading } = useSession();
	const router = useRouter();

	const [invitations, setInvitations] = useState<Invitation[]>([]);
	const [isLoading, setIsLoading] = useState(true);
	const [isSubmitting, setIsSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [success, setSuccess] = useState<string | null>(null);

	// Form state
	const [email, setEmail] = useState("");
	const [role, setRole] = useState<string>("USER");
	const [isDialogOpen, setIsDialogOpen] = useState(false);

	// Delete confirmation state
	const [deleteEmail, setDeleteEmail] = useState<string | null>(null);
	const [isDeleting, setIsDeleting] = useState(false);

	const fetchInvitations = useCallback(async () => {
		try {
			const response = await fetch("/api/invitations");
			if (!response.ok) {
				if (response.status === 403) {
					router.push("/settings");
					return;
				}
				throw new Error("Failed to load invitations");
			}
			const data = await response.json();
			setInvitations(data.invitations || []);
		} catch {
			setError("Error loading user list");
		} finally {
			setIsLoading(false);
		}
	}, [router]);

	useEffect(() => {
		if (!isSessionLoading && session?.user) {
			fetchInvitations();
		}
	}, [isSessionLoading, session, fetchInvitations]);

	// Check if user is admin
	useEffect(() => {
		if (!isSessionLoading && session?.user) {
			if (session.user.role !== "ADMIN") {
				router.push("/settings");
			}
		}
	}, [isSessionLoading, session, router]);

	const handleInvite = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsSubmitting(true);
		setError(null);
		setSuccess(null);

		try {
			const response = await fetch("/api/invitations", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ email, role }),
			});

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || "Failed to send invitation");
			}

			setSuccess(`Invitation sent to ${email}`);
			setEmail("");
			setRole("USER");
			setIsDialogOpen(false);
			fetchInvitations();
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Error sending invitation",
			);
		} finally {
			setIsSubmitting(false);
		}
	};

	const handleDelete = async (emailToDelete: string) => {
		setIsDeleting(true);
		setError(null);

		try {
			const response = await fetch(
				`/api/invitations?email=${encodeURIComponent(emailToDelete)}`,
				{ method: "DELETE" },
			);

			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || "Failed to remove user");
			}

			setSuccess(`User ${emailToDelete} removed successfully`);
			setDeleteEmail(null);
			fetchInvitations();
		} catch (err) {
			setError(
				err instanceof Error ? err.message : "Error removing user",
			);
		} finally {
			setIsDeleting(false);
		}
	};

	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString("en-US", {
			day: "2-digit",
			month: "short",
			year: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		});
	};

	if (isSessionLoading || isLoading) {
		return (
			<div className="space-y-6">
				<div>
					<Skeleton className="h-9 w-64 mb-2" />
					<Skeleton className="h-5 w-96" />
				</div>
				<Card>
					<CardHeader>
						<Skeleton className="h-6 w-48" />
					</CardHeader>
					<CardContent>
						<div className="space-y-4">
							{[1, 2, 3].map((i) => (
								<Skeleton key={i} className="h-16 w-full" />
							))}
						</div>
					</CardContent>
				</Card>
			</div>
		);
	}

	if (session?.user?.role !== "ADMIN") {
		return null;
	}

	const pendingInvitations = invitations.filter((inv) => !inv.acceptedAt);
	const activeUsers = invitations.filter((inv) => inv.acceptedAt);

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-3xl font-bold tracking-tight">
						User Management
					</h1>
					<p className="text-muted-foreground">
						Invite new users and manage platform access
					</p>
				</div>

				<Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
					<DialogTrigger asChild>
						<Button>
							<UserPlus className="mr-2 h-4 w-4" />
							Invite User
						</Button>
					</DialogTrigger>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>Invite New User</DialogTitle>
							<DialogDescription>
								Send an email invitation for a new user to
								access the platform.
							</DialogDescription>
						</DialogHeader>
						<form onSubmit={handleInvite} className="space-y-4">
							<div className="space-y-2">
								<Label htmlFor="email">Email</Label>
								<Input
									id="email"
									type="email"
									placeholder="user@example.com"
									value={email}
									onChange={(e) => setEmail(e.target.value)}
									disabled={isSubmitting}
									required
								/>
							</div>
							<div className="space-y-2">
								<Label htmlFor="role">Role</Label>
								<Select
									value={role}
									onValueChange={setRole}
									disabled={isSubmitting}
								>
									<SelectTrigger className="w-full">
										<SelectValue placeholder="Select a role" />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="USER">
											<div className="flex items-center gap-2">
												<Users className="h-4 w-4" />
												User
											</div>
										</SelectItem>
										<SelectItem value="ADMIN">
											<div className="flex items-center gap-2">
												<Shield className="h-4 w-4" />
												Administrator
											</div>
										</SelectItem>
									</SelectContent>
								</Select>
								<p className="text-xs text-muted-foreground">
									Administrators can invite and remove users.
								</p>
							</div>
							<DialogFooter>
								<Button
									type="button"
									variant="outline"
									onClick={() => setIsDialogOpen(false)}
									disabled={isSubmitting}
								>
									Cancel
								</Button>
								<Button
									type="submit"
									disabled={isSubmitting || !email}
								>
									{isSubmitting ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Sending...
										</>
									) : (
										<>
											<Mail className="mr-2 h-4 w-4" />
											Send Invitation
										</>
									)}
								</Button>
							</DialogFooter>
						</form>
					</DialogContent>
				</Dialog>
			</div>

			{error && (
				<div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3 text-sm text-destructive flex items-start gap-2">
					<AlertTriangle className="h-4 w-4 mt-0.5 shrink-0" />
					{error}
				</div>
			)}

			{success && (
				<div className="rounded-lg bg-primary/10 border border-primary/20 p-3 text-sm text-primary flex items-start gap-2">
					<CheckCircle2 className="h-4 w-4 mt-0.5 shrink-0" />
					{success}
				</div>
			)}

			{/* Statistics */}
			<div className="grid gap-4 md:grid-cols-3">
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">
							Total Users
						</CardTitle>
						<Users className="h-4 w-4 text-muted-foreground" />
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">
							{invitations.length}
						</div>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">
							Active Users
						</CardTitle>
						<CheckCircle2 className="h-4 w-4 text-muted-foreground" />
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">
							{activeUsers.length}
						</div>
					</CardContent>
				</Card>
				<Card>
					<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
						<CardTitle className="text-sm font-medium">
							Pending Invitations
						</CardTitle>
						<Clock className="h-4 w-4 text-muted-foreground" />
					</CardHeader>
					<CardContent>
						<div className="text-2xl font-bold">
							{pendingInvitations.length}
						</div>
					</CardContent>
				</Card>
			</div>

			{/* Users Table */}
			<Card>
				<CardHeader>
					<CardTitle>Authorized Users</CardTitle>
					<CardDescription>
						List of all users with platform access
					</CardDescription>
				</CardHeader>
				<CardContent>
					{invitations.length === 0 ? (
						<div className="text-center py-8 text-muted-foreground">
							<Users className="h-12 w-12 mx-auto mb-4 opacity-50" />
							<p>No users found</p>
							<p className="text-sm">
								Invite the first user by clicking the button
								above.
							</p>
						</div>
					) : (
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Email</TableHead>
									<TableHead>Name</TableHead>
									<TableHead>Role</TableHead>
									<TableHead>Status</TableHead>
									<TableHead>Invited at</TableHead>
									<TableHead className="text-right">
										Actions
									</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{invitations.map((invitation) => (
									<TableRow key={invitation.id}>
										<TableCell className="font-medium">
											{invitation.email}
										</TableCell>
										<TableCell>
											{invitation.user?.name || "-"}
										</TableCell>
										<TableCell>
											<Badge
												variant={
													invitation.role === "ADMIN"
														? "default"
														: "secondary"
												}
											>
												{invitation.role === "ADMIN" ? (
													<>
														<Shield className="mr-1 h-3 w-3" />
														Admin
													</>
												) : (
													<>
														<Users className="mr-1 h-3 w-3" />
														User
													</>
												)}
											</Badge>
										</TableCell>
										<TableCell>
											{invitation.acceptedAt ? (
												<Badge
													variant="outline"
													className="text-primary border-primary"
												>
													<CheckCircle2 className="mr-1 h-3 w-3" />
													Active
												</Badge>
											) : (
												<Badge variant="outline">
													<Clock className="mr-1 h-3 w-3" />
													Pending
												</Badge>
											)}
										</TableCell>
										<TableCell>
											{formatDate(invitation.invitedAt)}
										</TableCell>
										<TableCell className="text-right">
											{invitation.email !==
												session?.user?.email && (
												<Dialog
													open={
														deleteEmail ===
														invitation.email
													}
													onOpenChange={(open) =>
														!open &&
														setDeleteEmail(null)
													}
												>
													<DialogTrigger asChild>
														<Button
															variant="ghost"
															size="sm"
															className="text-destructive hover:text-destructive"
															onClick={() =>
																setDeleteEmail(
																	invitation.email,
																)
															}
														>
															<Trash2 className="h-4 w-4" />
														</Button>
													</DialogTrigger>
													<DialogContent>
														<DialogHeader>
															<DialogTitle>
																Remove User
															</DialogTitle>
															<DialogDescription>
																Are you sure you
																want to remove
																access for{" "}
																<strong>
																	{
																		invitation.email
																	}
																</strong>
																?
																{invitation.acceptedAt && (
																	<span className="block mt-2 text-destructive">
																		This
																		user
																		already
																		has an
																		active
																		account.
																		It will
																		be
																		permanently
																		removed.
																	</span>
																)}
															</DialogDescription>
														</DialogHeader>
														<DialogFooter>
															<Button
																variant="outline"
																onClick={() =>
																	setDeleteEmail(
																		null,
																	)
																}
																disabled={
																	isDeleting
																}
															>
																Cancel
															</Button>
															<Button
																variant="destructive"
																onClick={() =>
																	handleDelete(
																		invitation.email,
																	)
																}
																disabled={
																	isDeleting
																}
															>
																{isDeleting ? (
																	<>
																		<Loader2 className="mr-2 h-4 w-4 animate-spin" />
																		Removing...
																	</>
																) : (
																	<>
																		<Trash2 className="mr-2 h-4 w-4" />
																		Remove
																	</>
																)}
															</Button>
														</DialogFooter>
													</DialogContent>
												</Dialog>
											)}
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
