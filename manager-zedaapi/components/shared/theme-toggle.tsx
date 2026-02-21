"use client";

import { useCallback, useEffect, useState } from "react";
import { useTheme } from "next-themes";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";

export function ThemeToggle({ className }: { className?: string }) {
	const { resolvedTheme, setTheme } = useTheme();
	const [mounted, setMounted] = useState(false);

	useEffect(() => setMounted(true), []);

	const toggle = useCallback(() => {
		setTheme(resolvedTheme === "dark" ? "light" : "dark");
	}, [resolvedTheme, setTheme]);

	if (!mounted) {
		return (
			<button
				type="button"
				className={cn(
					"relative flex size-9 items-center justify-center rounded-full text-muted-foreground transition-colors hover:text-foreground",
					className,
				)}
				aria-label="Alternar tema"
				disabled
			>
				<div className="size-[18px]" />
			</button>
		);
	}

	const isDark = resolvedTheme === "dark";

	return (
		<button
			type="button"
			onClick={toggle}
			className={cn(
				"group relative flex size-9 items-center justify-center rounded-full text-muted-foreground transition-colors duration-200 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
				className,
			)}
			aria-label={isDark ? "Ativar modo claro" : "Ativar modo escuro"}
		>
			<AnimatePresence mode="wait" initial={false}>
				{isDark ? (
					<motion.svg
						key="sun"
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						strokeWidth={2}
						strokeLinecap="round"
						strokeLinejoin="round"
						className="size-[18px]"
						initial={{ rotate: -45, scale: 0, opacity: 0 }}
						animate={{ rotate: 0, scale: 1, opacity: 1 }}
						exit={{ rotate: 45, scale: 0, opacity: 0 }}
						transition={{
							duration: 0.25,
							ease: [0.22, 1, 0.36, 1],
						}}
					>
						<circle cx="12" cy="12" r="4" />
						<path d="M12 2v2" />
						<path d="M12 20v2" />
						<path d="m4.93 4.93 1.41 1.41" />
						<path d="m17.66 17.66 1.41 1.41" />
						<path d="M2 12h2" />
						<path d="M20 12h2" />
						<path d="m6.34 17.66-1.41 1.41" />
						<path d="m19.07 4.93-1.41 1.41" />
					</motion.svg>
				) : (
					<motion.svg
						key="moon"
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						strokeWidth={2}
						strokeLinecap="round"
						strokeLinejoin="round"
						className="size-[18px]"
						initial={{ rotate: 45, scale: 0, opacity: 0 }}
						animate={{ rotate: 0, scale: 1, opacity: 1 }}
						exit={{ rotate: -45, scale: 0, opacity: 0 }}
						transition={{
							duration: 0.25,
							ease: [0.22, 1, 0.36, 1],
						}}
					>
						<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
					</motion.svg>
				)}
			</AnimatePresence>
		</button>
	);
}
