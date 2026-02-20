"use client";

import { motion, useReducedMotion as _useReducedMotion } from "framer-motion";
import { fadeUp, staggerContainer } from "@/lib/design-tokens";

export function useReducedMotion() {
	return _useReducedMotion();
}

interface PageTransitionProps {
	children: React.ReactNode;
	className?: string;
}

export function PageTransition({ children, className }: PageTransitionProps) {
	const prefersReducedMotion = _useReducedMotion();

	if (prefersReducedMotion) {
		return <div className={className}>{children}</div>;
	}

	return (
		<motion.div
			initial={{ opacity: 0, y: 8 }}
			animate={{ opacity: 1, y: 0 }}
			transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
			className={className}
		>
			{children}
		</motion.div>
	);
}

interface AnimateInProps {
	children: React.ReactNode;
	delay?: number;
	className?: string;
}

export function AnimateIn({ children, delay = 0, className }: AnimateInProps) {
	const prefersReducedMotion = _useReducedMotion();

	if (prefersReducedMotion) {
		return <div className={className}>{children}</div>;
	}

	return (
		<motion.div
			custom={delay}
			variants={fadeUp}
			initial="hidden"
			animate="visible"
			className={className}
		>
			{children}
		</motion.div>
	);
}

interface StaggerGroupProps {
	children: React.ReactNode;
	className?: string;
}

export function StaggerGroup({ children, className }: StaggerGroupProps) {
	const prefersReducedMotion = _useReducedMotion();

	if (prefersReducedMotion) {
		return <div className={className}>{children}</div>;
	}

	return (
		<motion.div
			variants={staggerContainer}
			initial="hidden"
			animate="visible"
			className={className}
		>
			{children}
		</motion.div>
	);
}
