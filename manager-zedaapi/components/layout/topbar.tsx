"use client";

import { usePathname } from "next/navigation";
import { Bell, PanelLeft, Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
	Breadcrumb,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbPage,
	BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { useSidebarStore } from "@/stores/sidebar-store";
import { useAuth } from "@/hooks/use-auth";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { signOut } from "@/lib/auth-client";
import Link from "next/link";
import { Fragment } from "react";

const breadcrumbMap: Record<string, string> = {
	dashboard: "Dashboard",
	instances: "Instâncias",
	billing: "Assinatura",
	"api-keys": "Chaves API",
	organization: "Organização",
	members: "Membros",
	settings: "Configurações",
	profile: "Perfil",
	security: "Segurança",
	admin: "Admin",
	users: "Usuários",
	subscriptions: "Assinaturas",
	plans: "Planos",
	invoices: "Faturas",
	waitlist: "Waitlist",
	"feature-flags": "Feature Flags",
	"activity-log": "Log de Atividade",
};

interface TopbarProps {
	onMobileMenuToggle?: () => void;
}

export function Topbar({ onMobileMenuToggle }: TopbarProps) {
	const pathname = usePathname();
	const { toggle } = useSidebarStore();
	const { user, isAdmin } = useAuth();

	const segments = pathname.split("/").filter(Boolean);
	const breadcrumbs = segments.map((segment, index) => {
		const href = `/${segments.slice(0, index + 1).join("/")}`;
		const label = breadcrumbMap[segment] ?? segment;
		const isLast = index === segments.length - 1;
		return { href, label, isLast };
	});

	return (
		<header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-border/50 bg-background/80 px-4 backdrop-blur-sm lg:px-6">
			<Button
				variant="ghost"
				size="icon-sm"
				className="lg:hidden"
				onClick={onMobileMenuToggle}
			>
				<Menu className="size-4" />
			</Button>

			<Button
				variant="ghost"
				size="icon-sm"
				className="hidden lg:flex"
				onClick={toggle}
			>
				<PanelLeft className="size-4" />
			</Button>

			<Breadcrumb className="hidden sm:flex">
				<BreadcrumbList>
					{breadcrumbs.map((crumb, i) => (
						<Fragment key={crumb.href}>
							{i > 0 && <BreadcrumbSeparator />}
							<BreadcrumbItem>
								{crumb.isLast ? (
									<BreadcrumbPage>
										{crumb.label}
									</BreadcrumbPage>
								) : (
									<BreadcrumbLink asChild>
										<Link href={crumb.href}>
											{crumb.label}
										</Link>
									</BreadcrumbLink>
								)}
							</BreadcrumbItem>
						</Fragment>
					))}
				</BreadcrumbList>
			</Breadcrumb>

			<div className="ml-auto flex items-center gap-2">
				<Button variant="ghost" size="icon-sm" className="relative">
					<Bell className="size-4" />
					<span className="sr-only">Notificações</span>
				</Button>

				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button
							variant="ghost"
							className="relative size-8 rounded-full p-0"
						>
							<Avatar className="size-8 ring-2 ring-background">
								<AvatarFallback className="text-xs font-medium">
									{user?.name?.charAt(0)?.toUpperCase() ??
										"U"}
								</AvatarFallback>
							</Avatar>
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end" className="w-48">
						<div className="px-2 py-1.5">
							<p className="text-sm font-medium leading-none">
								{user?.name ?? "Usuário"}
							</p>
							<p className="mt-0.5 text-xs text-muted-foreground">
								{user?.email ?? ""}
							</p>
						</div>
						<DropdownMenuSeparator />
						<DropdownMenuItem asChild>
							<Link href="/perfil">Perfil</Link>
						</DropdownMenuItem>
						<DropdownMenuItem asChild>
							<Link href="/configuracoes">Configurações</Link>
						</DropdownMenuItem>
						{isAdmin && (
							<DropdownMenuItem asChild>
								<Link href="/admin">Painel Admin</Link>
							</DropdownMenuItem>
						)}
						<DropdownMenuSeparator />
						<DropdownMenuItem onClick={() => signOut()}>
							Sair
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</div>
		</header>
	);
}
