"use client";

import { useState } from "react";
import { TooltipProvider } from "@/components/ui/tooltip";
import { AdminSidebar } from "@/components/layout/admin-sidebar";
import { Topbar } from "@/components/layout/topbar";
import { MobileNav } from "@/components/layout/mobile-nav";

interface AdminShellProps {
	children: React.ReactNode;
}

export function AdminShell({ children }: AdminShellProps) {
	const [mobileOpen, setMobileOpen] = useState(false);

	return (
		<TooltipProvider delayDuration={0}>
			<div className="flex min-h-svh">
				<AdminSidebar />
				<div className="flex flex-1 flex-col overflow-hidden">
					<Topbar onMobileMenuToggle={() => setMobileOpen(true)} />
					<main className="flex-1 overflow-y-auto p-4 lg:p-6">
						<div className="mx-auto max-w-7xl">{children}</div>
					</main>
				</div>
				<MobileNav open={mobileOpen} onOpenChange={setMobileOpen} />
			</div>
		</TooltipProvider>
	);
}
