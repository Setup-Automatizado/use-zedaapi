"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
	Sheet,
	SheetTrigger,
	SheetContent,
	SheetHeader,
	SheetTitle,
	SheetClose,
} from "@/components/ui/sheet";
import { MenuIcon, ArrowRightIcon } from "lucide-react";
import { ThemeToggle } from "@/components/shared/theme-toggle";

const navLinks = [
	{ label: "Recursos", href: "#recursos" },
	{ label: "Como Funciona", href: "#como-funciona" },
	{ label: "Preços", href: "#precos" },
	{ label: "Integrações", href: "#integracoes" },
	{ label: "FAQ", href: "#faq" },
	{ label: "Contato", href: "#contato" },
] as const;

export function Header() {
	const [scrolled, setScrolled] = useState(false);
	const [activeSection, setActiveSection] = useState("");

	useEffect(() => {
		function onScroll() {
			setScrolled(window.scrollY > 16);
		}
		onScroll();
		window.addEventListener("scroll", onScroll, { passive: true });
		return () => window.removeEventListener("scroll", onScroll);
	}, []);

	// Track active section via Intersection Observer
	useEffect(() => {
		const ids = navLinks.map((l) => l.href.replace("#", ""));
		const observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					if (entry.isIntersecting) {
						setActiveSection(entry.target.id);
					}
				}
			},
			{ rootMargin: "-40% 0px -55% 0px", threshold: 0 },
		);

		for (const id of ids) {
			const el = document.getElementById(id);
			if (el) observer.observe(el);
		}

		return () => observer.disconnect();
	}, []);

	const handleSmoothScroll = useCallback(
		(e: React.MouseEvent<HTMLAnchorElement>, href: string) => {
			e.preventDefault();
			const id = href.replace("#", "");
			const el = document.getElementById(id);
			if (el) {
				el.scrollIntoView({ behavior: "smooth", block: "start" });
				window.history.replaceState(null, "", href);
			}
		},
		[],
	);

	return (
		<header
			className={cn(
				"sticky top-0 z-50 w-full transition-all duration-500",
				scrolled
					? "bg-background/70 backdrop-blur-2xl border-b border-border/40 shadow-[0_1px_3px_rgba(0,0,0,0.04)]"
					: "bg-transparent",
			)}
		>
			<div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
				{/* Logo */}
				<Link
					href="/"
					className="group relative flex items-center gap-2.5 transition-all duration-300"
				>
					{/* Logo image with glow effect */}
					<div className="relative">
						<div
							aria-hidden="true"
							className="pointer-events-none absolute inset-0 rounded-full bg-primary/20 blur-md opacity-0 transition-opacity duration-300 group-hover:opacity-100"
						/>
						<Image
							src="/favicon-96x96.png"
							alt="Zé da API"
							width={36}
							height={36}
							className="relative size-9 rounded-xl object-contain transition-transform duration-300 group-hover:scale-110"
							priority
						/>
					</div>
					<div className="flex flex-col">
						<span className="text-base font-bold tracking-tight text-foreground leading-none">
							Zé da API
						</span>
						<span className="text-[10px] font-medium text-muted-foreground/70 tracking-wider uppercase leading-none mt-0.5">
							WhatsApp API
						</span>
					</div>
				</Link>

				{/* Desktop Nav - pill-style with active indicator */}
				<nav className="hidden items-center rounded-full border border-border/50 bg-muted/30 px-1 py-1 md:flex">
					{navLinks.map((link) => {
						const isActive =
							activeSection === link.href.replace("#", "");
						return (
							<a
								key={link.href}
								href={link.href}
								onClick={(e) =>
									handleSmoothScroll(e, link.href)
								}
								className={cn(
									"relative rounded-full px-3.5 py-1.5 text-[13px] font-medium transition-all duration-200",
									isActive
										? "text-foreground"
										: "text-muted-foreground hover:text-foreground",
								)}
							>
								{isActive && (
									<motion.div
										layoutId="nav-active-pill"
										className="absolute inset-0 rounded-full bg-background shadow-sm ring-1 ring-border/50"
										transition={{
											type: "spring",
											bounce: 0.15,
											duration: 0.5,
										}}
									/>
								)}
								<span className="relative z-10">
									{link.label}
								</span>
							</a>
						);
					})}
				</nav>

				{/* Desktop CTA */}
				<div className="hidden items-center gap-2 md:flex">
					<ThemeToggle />
					<Button
						variant="ghost"
						size="sm"
						asChild
						className="text-[13px] font-medium text-muted-foreground hover:text-foreground"
					>
						<Link href="/login">Entrar</Link>
					</Button>
					<Button
						size="sm"
						asChild
						className="group/cta relative h-9 rounded-full px-5 text-[13px] font-semibold shadow-md shadow-primary/20 transition-all duration-300 hover:shadow-lg hover:shadow-primary/30"
					>
						<Link href="/cadastro">
							Criar Conta Grátis
							<ArrowRightIcon className="size-3.5 ml-1 transition-transform duration-200 group-hover/cta:translate-x-0.5" />
						</Link>
					</Button>
				</div>

				{/* Mobile Menu */}
				<Sheet>
					<SheetTrigger asChild>
						<Button
							variant="ghost"
							size="icon-sm"
							className="md:hidden"
							aria-label="Abrir menu"
						>
							<MenuIcon className="size-5" />
						</Button>
					</SheetTrigger>
					<SheetContent side="right" className="w-80">
						<SheetHeader>
							<SheetTitle className="flex items-center gap-2.5">
								<Image
									src="/favicon-96x96.png"
									alt="Zé da API"
									width={32}
									height={32}
									className="size-8 rounded-xl object-contain"
								/>
								<div className="flex flex-col">
									<span className="text-base font-bold tracking-tight leading-none">
										Zé da API
									</span>
									<span className="text-[10px] font-medium text-muted-foreground/70 tracking-wider uppercase leading-none mt-0.5">
										WhatsApp API
									</span>
								</div>
							</SheetTitle>
						</SheetHeader>
						<div className="flex items-center justify-between px-6 pt-4">
							<span className="text-xs font-medium uppercase tracking-wider text-muted-foreground/60">
								Menu
							</span>
							<ThemeToggle />
						</div>
						<nav className="flex flex-col gap-1 px-6 pt-2">
							{navLinks.map((link) => (
								<SheetClose key={link.href} asChild>
									<a
										href={link.href}
										className={cn(
											"rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-200",
											activeSection ===
												link.href.replace("#", "")
												? "text-foreground bg-muted/80"
												: "text-muted-foreground hover:text-foreground hover:bg-muted/60",
										)}
									>
										{link.label}
									</a>
								</SheetClose>
							))}
						</nav>
						<div className="mt-auto flex flex-col gap-2.5 p-6">
							<Button
								variant="outline"
								asChild
								className="w-full rounded-xl"
							>
								<Link href="/login">Entrar</Link>
							</Button>
							<Button
								asChild
								className="w-full rounded-xl shadow-md shadow-primary/20"
							>
								<Link href="/cadastro">
									Criar Conta Grátis
									<ArrowRightIcon
										className="size-3.5 ml-1"
										data-icon="inline-end"
									/>
								</Link>
							</Button>
						</div>
					</SheetContent>
				</Sheet>
			</div>
		</header>
	);
}
