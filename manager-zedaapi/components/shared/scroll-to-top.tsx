"use client";

import { useCallback, useEffect, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";

export function ScrollToTop() {
	const [visible, setVisible] = useState(false);

	useEffect(() => {
		function onScroll() {
			setVisible(window.scrollY > 400);
		}
		onScroll();
		window.addEventListener("scroll", onScroll, { passive: true });
		return () => window.removeEventListener("scroll", onScroll);
	}, []);

	const scrollToTop = useCallback(() => {
		window.scrollTo({ top: 0, behavior: "smooth" });
	}, []);

	return (
		<AnimatePresence>
			{visible && (
				<motion.button
					type="button"
					onClick={scrollToTop}
					initial={{ opacity: 0, scale: 0.8, y: 10 }}
					animate={{ opacity: 1, scale: 1, y: 0 }}
					exit={{ opacity: 0, scale: 0.8, y: 10 }}
					transition={{
						type: "spring",
						stiffness: 400,
						damping: 25,
					}}
					whileHover={{ scale: 1.1 }}
					whileTap={{ scale: 0.9 }}
					className="fixed bottom-24 right-6 z-50 flex size-11 items-center justify-center rounded-full border border-border/50 bg-background/80 text-muted-foreground shadow-lg shadow-black/5 backdrop-blur-xl transition-colors duration-200 hover:border-border hover:text-foreground hover:shadow-xl dark:shadow-black/20"
					aria-label="Voltar ao topo"
				>
					<motion.svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						strokeWidth={2}
						strokeLinecap="round"
						strokeLinejoin="round"
						className="size-4"
						animate={{ y: [0, -2, 0] }}
						transition={{
							duration: 1.5,
							repeat: Infinity,
							ease: "easeInOut",
						}}
					>
						<path d="m18 15-6-6-6 6" />
					</motion.svg>
				</motion.button>
			)}
		</AnimatePresence>
	);
}
