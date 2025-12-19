"use client";

import Link from "next/link";
import { useSession } from "@/lib/auth-client";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ShieldCheck, User, Users } from "lucide-react";

interface SettingsSection {
	title: string;
	description: string;
	icon: typeof ShieldCheck;
	href: string;
	disabled?: boolean;
	adminOnly?: boolean;
}

const settingsSections: SettingsSection[] = [
	{
		title: "Security",
		description: "Manage two-factor authentication and password",
		icon: ShieldCheck,
		href: "/settings/security",
	},
	{
		title: "User Management",
		description: "Invite users and manage access",
		icon: Users,
		href: "/settings/admin",
		adminOnly: true,
	},
	{
		title: "Profile",
		description: "Update your personal information",
		icon: User,
		href: "/settings/profile",
	},
];

export default function SettingsPage() {
	const { data: session } = useSession();
	const isAdmin = session?.user?.role === "ADMIN";

	const visibleSections = settingsSections.filter(
		(section) => !section.adminOnly || isAdmin,
	);

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-3xl font-bold tracking-tight">Settings</h1>
				<p className="text-muted-foreground">
					Manage your account settings and preferences
				</p>
			</div>

			<div className="grid gap-4 md:grid-cols-2">
				{visibleSections.map((section) => {
					const Icon = section.icon;
					return (
						<Card
							key={section.title}
							className={section.disabled ? "opacity-60" : ""}
						>
							<CardHeader>
								<div className="flex items-center gap-3">
									<div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
										<Icon className="h-5 w-5 text-primary" />
									</div>
									<div>
										<CardTitle className="text-lg">
											{section.title}
										</CardTitle>
										<CardDescription>
											{section.description}
										</CardDescription>
									</div>
								</div>
							</CardHeader>
							<CardContent>
								{section.disabled ? (
									<Button variant="outline" disabled>
										Coming soon
									</Button>
								) : (
									<Button variant="outline" asChild>
										<Link href={section.href}>Manage</Link>
									</Button>
								)}
							</CardContent>
						</Card>
					);
				})}
			</div>
		</div>
	);
}
