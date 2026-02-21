"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
	Sheet,
	SheetContent,
	SheetHeader,
	SheetTitle,
} from "@/components/ui/sheet";
import {
	LayoutDashboard,
	Smartphone,
	CreditCard,
	Settings,
	Key,
	Building2,
	Users,
	LogOut,
} from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/shared/theme-toggle";
import { signOut } from "@/lib/auth-client";
import { useAuth } from "@/hooks/use-auth";

const mainNav = [
	{ name: "Painel", href: "/painel", icon: LayoutDashboard },
	{ name: "Instâncias", href: "/instancias", icon: Smartphone },
	{ name: "Assinatura", href: "/faturamento", icon: CreditCard },
	{ name: "Chaves API", href: "/chaves-api", icon: Key },
];

const orgNav = [
	{ name: "Organização", href: "/organizacao", icon: Building2 },
	{ name: "Membros", href: "/organizacao/membros", icon: Users },
	{ name: "Configurações", href: "/configuracoes", icon: Settings },
];

interface MobileNavProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function MobileNav({ open, onOpenChange }: MobileNavProps) {
	const pathname = usePathname();
	const { user } = useAuth();

	function renderNavItem(item: (typeof mainNav)[number]) {
		const isActive =
			pathname === item.href ||
			(item.href !== "/painel" && pathname.startsWith(item.href));

		return (
			<Link
				key={item.name}
				href={item.href}
				onClick={() => onOpenChange(false)}
				className={cn(
					"flex h-9 items-center gap-3 rounded-lg px-3 text-sm transition-colors duration-150",
					isActive
						? "bg-accent text-accent-foreground font-medium"
						: "text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground",
				)}
			>
				<item.icon className="size-4 shrink-0" />
				<span>{item.name}</span>
			</Link>
		);
	}

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
			<SheetContent side="left" className="w-[280px] p-0">
				<SheetHeader className="border-b border-border/50 px-4 py-4">
					<SheetTitle className="flex items-center gap-2">
						<div className="flex size-8 items-center justify-center rounded-xl bg-primary text-primary-foreground text-sm font-bold">
							Z
						</div>
						<span className="text-sm font-semibold tracking-tight">
							Zé da API
						</span>
					</SheetTitle>
				</SheetHeader>

				<nav className="flex-1 overflow-y-auto p-3">
					<div className="space-y-1">
						{mainNav.map(renderNavItem)}
					</div>

					<p className="mb-2 mt-6 px-3 text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
						Organização
					</p>
					<div className="space-y-1">{orgNav.map(renderNavItem)}</div>
				</nav>

				<div className="mt-auto border-t border-border/50 p-3">
					<div className="flex items-center justify-between mb-3">
						<ThemeToggle />
					</div>
					<Separator className="mb-3" />
					<div className="flex items-center gap-3">
						<div className="flex size-8 items-center justify-center rounded-full bg-muted text-xs font-medium">
							{user?.name?.charAt(0)?.toUpperCase() ?? "U"}
						</div>
						<div className="flex-1 min-w-0">
							<p className="truncate text-sm font-medium leading-none">
								{user?.name ?? "Usuário"}
							</p>
							<p className="truncate text-xs text-muted-foreground mt-0.5">
								{user?.email ?? ""}
							</p>
						</div>
						<Button
							variant="ghost"
							size="icon-sm"
							onClick={() => signOut()}
							className="text-muted-foreground shrink-0"
						>
							<LogOut className="size-4" />
						</Button>
					</div>
				</div>
			</SheetContent>
		</Sheet>
	);
}
