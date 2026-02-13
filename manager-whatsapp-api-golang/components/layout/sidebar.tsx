"use client";

import {
	Activity,
	BarChart3,
	Globe,
	LayoutDashboard,
	LogOut,
	Settings,
	Smartphone,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as React from "react";
import { ApiVersion } from "@/components/layout/api-version";
import { Button } from "@/components/ui/button";
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { useSidebar } from "@/lib/sidebar-context";
import { cn } from "@/lib/utils";

interface NavItem {
	name: string;
	href: string;
	icon: React.ElementType;
}

const navItems: NavItem[] = [
	{ name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
	{ name: "Instances", href: "/instances", icon: Smartphone },
	{ name: "Proxy Pool", href: "/proxy-pool", icon: Globe },
	{ name: "Health", href: "/health", icon: Activity },
	{ name: "Metrics", href: "/metrics", icon: BarChart3 },
	{ name: "Settings", href: "/settings", icon: Settings },
];

interface SidebarProps {
	user?: {
		name?: string | null;
		email?: string | null;
		image?: string | null;
	};
	onSignOut?: () => void;
}

export function Sidebar({ user, onSignOut }: SidebarProps) {
	const pathname = usePathname();
	const { isCollapsed } = useSidebar();

	const isActive = (href: string) => {
		if (href === "/dashboard") {
			return pathname === "/dashboard" || pathname === "/";
		}
		return pathname.startsWith(href);
	};

	return (
		<TooltipProvider delayDuration={0}>
			<div className="flex h-full flex-col border-r bg-background">
				{/* Header with Logo */}
				<div className="flex h-14 items-center px-3">
					<Link
						href="/"
						className={cn(
							"flex items-center",
							isCollapsed
								? "justify-center w-full"
								: "space-x-2 flex-1 px-3",
						)}
					>
						<Image
							src="/android-chrome-96x96.png"
							alt="WhatsApp API"
							width={28}
							height={28}
							className="flex-shrink-0 rounded"
						/>
						{!isCollapsed && (
							<span className="font-bold text-lg">
								WhatsApp API
							</span>
						)}
					</Link>
				</div>

				{/* Navigation */}
				<nav className="flex-1 space-y-1 px-3 py-4">
					{navItems.map((item) => {
						const Icon = item.icon;
						const active = isActive(item.href);

						const linkContent = (
							<Link
								href={item.href}
								className={cn(
									"flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
									active
										? "bg-primary text-primary-foreground"
										: "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
									isCollapsed && "justify-center px-2",
								)}
							>
								<Icon className="h-4 w-4 flex-shrink-0" />
								{!isCollapsed && <span>{item.name}</span>}
							</Link>
						);

						if (isCollapsed) {
							return (
								<Tooltip key={item.href}>
									<TooltipTrigger asChild>
										{linkContent}
									</TooltipTrigger>
									<TooltipContent
										side="right"
										sideOffset={10}
									>
										{item.name}
									</TooltipContent>
								</Tooltip>
							);
						}

						return (
							<React.Fragment key={item.href}>
								{linkContent}
							</React.Fragment>
						);
					})}
				</nav>

				{/* User Section */}
				<div className="p-4 space-y-2">
					{user && !isCollapsed && (
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
						<>
							{isCollapsed ? (
								<Tooltip>
									<TooltipTrigger asChild>
										<Button
											variant="ghost"
											size="icon"
											className="w-full"
											onClick={onSignOut}
										>
											<LogOut className="h-4 w-4" />
										</Button>
									</TooltipTrigger>
									<TooltipContent
										side="right"
										sideOffset={10}
									>
										Sign out
									</TooltipContent>
								</Tooltip>
							) : (
								<Button
									variant="ghost"
									className="w-full justify-start gap-3"
									onClick={onSignOut}
								>
									<LogOut className="h-4 w-4" />
									<span>Sign out</span>
								</Button>
							)}
						</>
					)}

					{/* API Version */}
					<ApiVersion collapsed={isCollapsed} className="pt-2" />
				</div>
			</div>
		</TooltipProvider>
	);
}
