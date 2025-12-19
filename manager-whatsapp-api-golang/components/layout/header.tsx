"use client";

import * as React from "react";
import { Menu, PanelLeftClose, PanelLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "./theme-toggle";
import { UserNav } from "./user-nav";
import { DynamicBreadcrumb } from "./breadcrumb";
import { useSidebar } from "@/lib/sidebar-context";

interface HeaderProps {
	user?: {
		name?: string | null;
		email?: string | null;
		image?: string | null;
	};
	onMobileMenuToggle?: () => void;
}

export function Header({ user, onMobileMenuToggle }: HeaderProps) {
	const { isCollapsed, toggle } = useSidebar();

	return (
		<header className="sticky top-0 z-40 flex h-14 items-center bg-background">
			{/* Mobile menu button */}
			<Button
				variant="ghost"
				size="icon"
				className="md:hidden ml-4"
				onClick={onMobileMenuToggle}
			>
				<Menu className="h-5 w-5" />
				<span className="sr-only">Toggle menu</span>
			</Button>

			{/* Sidebar collapse toggle - positioned near the border */}
			<Button
				variant="ghost"
				size="icon"
				onClick={toggle}
				className="hidden md:flex h-8 w-8 ml-2"
				aria-label={isCollapsed ? "Expand sidebar" : "Collapse sidebar"}
			>
				{isCollapsed ? (
					<PanelLeft className="h-4 w-4" />
				) : (
					<PanelLeftClose className="h-4 w-4" />
				)}
			</Button>

			{/* Breadcrumb */}
			<div className="flex-1 ml-2">
				<DynamicBreadcrumb />
			</div>

			{/* Right side actions */}
			<div className="flex items-center gap-2 mr-6">
				<ThemeToggle />
				<UserNav user={user} />
			</div>
		</header>
	);
}
