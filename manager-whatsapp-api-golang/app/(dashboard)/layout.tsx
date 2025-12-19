"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { MobileNav } from "@/components/layout/mobile-nav";
import { SidebarProvider, useSidebar } from "@/lib/sidebar-context";
import { signOut, useSession } from "@/lib/auth-client";

function DashboardContent({ children }: { children: React.ReactNode }) {
	const router = useRouter();
	const [mobileNavOpen, setMobileNavOpen] = React.useState(false);
	const { data: session, isPending } = useSession();
	const { isCollapsed } = useSidebar();

	// Redirect to login if not authenticated
	React.useEffect(() => {
		if (!isPending && !session?.user) {
			router.push("/login");
		}
	}, [isPending, session, router]);

	const user = session?.user
		? {
				name: session.user.name,
				email: session.user.email,
				image: session.user.image,
			}
		: undefined;

	const handleSignOut = async () => {
		try {
			await signOut({
				fetchOptions: {
					onSuccess: () => {
						router.push("/login");
						router.refresh();
					},
				},
			});
		} catch (error) {
			console.error("Failed to sign out:", error);
		}
	};

	return (
		<div className="min-h-screen bg-muted/40">
			<aside
				className={`hidden md:flex md:flex-col md:fixed md:inset-y-0 md:z-50 transition-all duration-300 ${
					isCollapsed ? "md:w-16" : "md:w-64"
				}`}
			>
				<Sidebar user={user} onSignOut={handleSignOut} />
			</aside>

			<MobileNav
				open={mobileNavOpen}
				onOpenChange={setMobileNavOpen}
				user={user}
				onSignOut={handleSignOut}
			/>

			<div
				className={`transition-all duration-300 ${
					isCollapsed ? "md:pl-16" : "md:pl-64"
				}`}
			>
				<Header
					user={user}
					onMobileMenuToggle={() => setMobileNavOpen(true)}
				/>

				<main className="p-6">{children}</main>
			</div>
		</div>
	);
}

export default function DashboardLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<SidebarProvider>
			<DashboardContent>{children}</DashboardContent>
		</SidebarProvider>
	);
}
