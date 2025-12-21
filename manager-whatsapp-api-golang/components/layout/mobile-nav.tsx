"use client";

import {
	Activity,
	BarChart3,
	LayoutDashboard,
	LogOut,
	Settings,
	Smartphone,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type * as React from "react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
	Sheet,
	SheetContent,
	SheetHeader,
	SheetTitle,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

interface NavItem {
	name: string;
	href: string;
	icon: React.ElementType;
}

const navItems: NavItem[] = [
	{ name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
	{ name: "Instances", href: "/instances", icon: Smartphone },
	{ name: "Health", href: "/health", icon: Activity },
	{ name: "Metrics", href: "/metrics", icon: BarChart3 },
	{ name: "Settings", href: "/settings", icon: Settings },
];

interface MobileNavProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	user?: {
		name?: string | null;
		email?: string | null;
		image?: string | null;
	};
	onSignOut?: () => void;
}

export function MobileNav({
	open,
	onOpenChange,
	user,
	onSignOut,
}: MobileNavProps) {
	const pathname = usePathname();

	const isActive = (href: string) => {
		if (href === "/dashboard") {
			return pathname === "/dashboard" || pathname === "/";
		}
		return pathname.startsWith(href);
	};

	const handleLinkClick = () => {
		onOpenChange(false);
	};

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
			<SheetContent side="left" className="w-64 p-0">
				<SheetHeader className="flex h-14 items-center border-b px-6">
					<SheetTitle className="flex items-center space-x-2">
						<Smartphone className="h-6 w-6" />
						<span className="font-bold text-lg">WhatsApp API</span>
					</SheetTitle>
				</SheetHeader>

				<nav className="flex-1 space-y-1 px-3 py-4">
					{navItems.map((item) => {
						const Icon = item.icon;
						const active = isActive(item.href);

						return (
							<Link
								key={item.href}
								href={item.href}
								onClick={handleLinkClick}
								className={cn(
									"flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
									active
										? "bg-primary text-primary-foreground"
										: "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
								)}
							>
								<Icon className="h-4 w-4" />
								<span>{item.name}</span>
							</Link>
						);
					})}
				</nav>

				<Separator />

				<div className="p-4 space-y-2">
					{user && (
						<div className="px-3 py-2">
							<p className="text-sm font-medium truncate">
								{user.name || user.email || "User"}
							</p>
							{user.email && user.name && (
								<p className="text-xs text-muted-foreground truncate">
									{user.email}
								</p>
							)}
						</div>
					)}

					{onSignOut && (
						<Button
							variant="ghost"
							className="w-full justify-start gap-3"
							onClick={() => {
								onSignOut();
								handleLinkClick();
							}}
						>
							<LogOut className="h-4 w-4" />
							<span>Sign out</span>
						</Button>
					)}
				</div>
			</SheetContent>
		</Sheet>
	);
}
