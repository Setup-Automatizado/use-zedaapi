"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { useSidebarStore } from "@/stores/sidebar-store";
import {
	LayoutDashboard,
	Users,
	CreditCard,
	Smartphone,
	Receipt,
	Clock,
	Settings,
	Flag,
	Activity,
	ChevronLeft,
	LogOut,
	ArrowLeft,
	Layers,
	UserCheck,
	BookOpen,
	HelpCircle,
	BookA,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { ThemeToggle } from "@/components/shared/theme-toggle";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { signOut } from "@/lib/auth-client";
import { useAuth } from "@/hooks/use-auth";

const mainNav = [
	{ name: "Painel Admin", href: "/admin", icon: LayoutDashboard },
	{ name: "Usuarios", href: "/admin/usuarios", icon: Users },
	{ name: "Assinaturas", href: "/admin/assinaturas", icon: CreditCard },
	{ name: "Planos", href: "/admin/planos", icon: Layers },
	{ name: "Instancias", href: "/admin/instancias", icon: Smartphone },
	{ name: "Faturas", href: "/admin/faturas", icon: Receipt },
];

const contentNav = [
	{ name: "Blog", href: "/admin/blog", icon: BookOpen },
	{ name: "Suporte", href: "/admin/suporte", icon: HelpCircle },
	{ name: "Glossario", href: "/admin/glossario", icon: BookA },
];

const systemNav = [
	{ name: "Lista de Espera", href: "/admin/lista-de-espera", icon: Clock },
	{ name: "Afiliados", href: "/admin/afiliados", icon: UserCheck },
	{ name: "Funcionalidades", href: "/admin/funcionalidades", icon: Flag },
	{
		name: "Log de Atividade",
		href: "/admin/registro-atividades",
		icon: Activity,
	},
	{ name: "Configuracoes", href: "/admin/configuracoes", icon: Settings },
];

export function AdminSidebar() {
	const pathname = usePathname();
	const { isCollapsed, toggle } = useSidebarStore();
	const { user } = useAuth();

	function renderNavItem(item: (typeof mainNav)[number]) {
		const isActive =
			pathname === item.href ||
			(item.href !== "/admin" && pathname.startsWith(item.href));

		const link = (
			<Link
				key={item.name}
				href={item.href}
				className={cn(
					"flex h-9 items-center gap-3 rounded-lg px-3 text-sm transition-colors duration-150",
					isActive
						? "bg-accent text-accent-foreground font-medium"
						: "text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground",
					isCollapsed && "justify-center px-0",
				)}
			>
				<item.icon className="size-4 shrink-0" />
				{!isCollapsed && <span>{item.name}</span>}
			</Link>
		);

		if (isCollapsed) {
			return (
				<Tooltip key={item.name}>
					<TooltipTrigger asChild>{link}</TooltipTrigger>
					<TooltipContent side="right" sideOffset={8}>
						{item.name}
					</TooltipContent>
				</Tooltip>
			);
		}

		return link;
	}

	return (
		<aside
			className={cn(
				"sticky top-0 hidden h-svh flex-col border-r border-border/50 bg-sidebar transition-all duration-200 lg:flex",
				isCollapsed ? "w-16" : "w-[240px]",
			)}
		>
			{/* Logo */}
			<div
				className={cn(
					"flex h-14 items-center border-b border-border/50 px-3",
					isCollapsed ? "justify-center" : "justify-between",
				)}
			>
				{!isCollapsed && (
					<Link href="/admin" className="flex items-center gap-2">
						<div className="flex size-8 items-center justify-center rounded-xl bg-primary text-primary-foreground text-sm font-bold">
							Z
						</div>
						<div className="flex items-baseline gap-1">
							<span className="text-sm font-semibold tracking-tight">
								Zé da API
							</span>
							<span className="text-[11px] text-muted-foreground font-medium">
								Admin
							</span>
						</div>
					</Link>
				)}
				{isCollapsed && (
					<Link href="/admin">
						<div className="flex size-8 items-center justify-center rounded-xl bg-primary text-primary-foreground text-sm font-bold">
							Z
						</div>
					</Link>
				)}
				{!isCollapsed && (
					<Button
						variant="ghost"
						size="icon-sm"
						onClick={toggle}
						className="text-muted-foreground"
					>
						<ChevronLeft className="size-4" />
					</Button>
				)}
			</div>

			{/* Back to Painel */}
			{!isCollapsed && (
				<div className="px-3 pt-3">
					<Link href="/painel">
						<Button
							variant="ghost"
							size="sm"
							className="w-full justify-start gap-2 text-muted-foreground"
						>
							<ArrowLeft className="size-4" />
							Voltar ao Painel
						</Button>
					</Link>
				</div>
			)}

			{/* Navigation */}
			<nav className="flex-1 overflow-y-auto p-3">
				<div className="space-y-1">{mainNav.map(renderNavItem)}</div>

				{!isCollapsed && (
					<p className="mb-2 mt-6 px-3 text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
						Conteudo
					</p>
				)}
				{isCollapsed && <Separator className="my-3" />}
				<div className="space-y-1">{contentNav.map(renderNavItem)}</div>

				{!isCollapsed && (
					<p className="mb-2 mt-6 px-3 text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
						Sistema
					</p>
				)}
				{isCollapsed && <Separator className="my-3" />}
				<div className="space-y-1">{systemNav.map(renderNavItem)}</div>
			</nav>

			{/* Bottom section */}
			<div className="mt-auto border-t border-border/50 p-3">
				<div
					className={cn(
						"flex items-center gap-2",
						isCollapsed ? "flex-col" : "justify-between",
					)}
				>
					<ThemeToggle />
				</div>
				<Separator className="my-3" />
				<div
					className={cn(
						"flex items-center gap-3",
						isCollapsed && "justify-center",
					)}
				>
					<div className="flex size-8 items-center justify-center rounded-full bg-destructive/10 text-destructive text-xs font-medium">
						{user?.name?.charAt(0)?.toUpperCase() ?? "A"}
					</div>
					{!isCollapsed && (
						<div className="flex-1 min-w-0">
							<p className="truncate text-sm font-medium leading-none">
								{user?.name ?? "Admin"}
							</p>
							<p className="truncate text-xs text-muted-foreground mt-0.5">
								Administrador
							</p>
						</div>
					)}
					{!isCollapsed && (
						<Tooltip>
							<TooltipTrigger asChild>
								<Button
									variant="ghost"
									size="icon-sm"
									onClick={() => signOut()}
									className="text-muted-foreground shrink-0"
								>
									<LogOut className="size-4" />
								</Button>
							</TooltipTrigger>
							<TooltipContent side="top">Sair</TooltipContent>
						</Tooltip>
					)}
				</div>
			</div>
		</aside>
	);
}
