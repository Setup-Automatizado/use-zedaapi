"use client";

import { motion } from "motion/react";
import { cn } from "@/lib/utils";

interface LoaderProps extends React.HTMLAttributes<HTMLDivElement> {
  title?: string;
  subtitle?: string;
  size?: "sm" | "md" | "lg";
  showText?: boolean;
}

function Loader({
  title = "Carregando...",
  subtitle = "Aguarde enquanto preparamos tudo para você",
  size = "md",
  showText = true,
  className,
  ...props
}: LoaderProps) {
  const sizeConfig = {
    sm: {
      container: "size-16",
      titleClass: "text-sm/tight font-medium",
      subtitleClass: "text-xs/relaxed",
      spacing: "space-y-2",
      maxWidth: "max-w-48",
    },
    md: {
      container: "size-24",
      titleClass: "text-base/snug font-medium",
      subtitleClass: "text-sm/relaxed",
      spacing: "space-y-3",
      maxWidth: "max-w-56",
    },
    lg: {
      container: "size-32",
      titleClass: "text-lg/tight font-semibold",
      subtitleClass: "text-base/relaxed",
      spacing: "space-y-4",
      maxWidth: "max-w-64",
    },
  };

  const config = sizeConfig[size];

  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-6 p-8",
        className
      )}
      {...props}
    >
      <motion.div
        className={cn("relative", config.container)}
        animate={{
          scale: [1, 1.02, 1],
        }}
        transition={{
          duration: 4,
          repeat: Number.POSITIVE_INFINITY,
          ease: [0.4, 0, 0.6, 1],
        }}
      >
        {/* Outer elegant ring - Light mode (primary/lime color) */}
        <motion.div
          className="absolute inset-0 rounded-full dark:hidden"
          style={{
            background: `conic-gradient(from 0deg, transparent 0deg, hsl(var(--primary)) 90deg, transparent 180deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 35%, black 37%, black 39%, transparent 41%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 35%, black 37%, black 39%, transparent 41%)`,
            opacity: 0.8,
          }}
          animate={{
            rotate: [0, 360],
          }}
          transition={{
            duration: 3,
            repeat: Number.POSITIVE_INFINITY,
            ease: "linear",
          }}
        />

        {/* Primary animated ring - Light mode */}
        <motion.div
          className="absolute inset-0 rounded-full dark:hidden"
          style={{
            background: `conic-gradient(from 0deg, transparent 0deg, hsl(var(--primary)) 120deg, hsl(var(--primary) / 0.5) 240deg, transparent 360deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 42%, black 44%, black 48%, transparent 50%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 42%, black 44%, black 48%, transparent 50%)`,
            opacity: 0.9,
          }}
          animate={{
            rotate: [0, 360],
          }}
          transition={{
            duration: 2.5,
            repeat: Number.POSITIVE_INFINITY,
            ease: [0.4, 0, 0.6, 1],
          }}
        />

        {/* Secondary elegant ring - counter rotation - Light mode */}
        <motion.div
          className="absolute inset-0 rounded-full dark:hidden"
          style={{
            background: `conic-gradient(from 180deg, transparent 0deg, hsl(var(--primary) / 0.6) 45deg, transparent 90deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 52%, black 54%, black 56%, transparent 58%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 52%, black 54%, black 56%, transparent 58%)`,
            opacity: 0.35,
          }}
          animate={{
            rotate: [0, -360],
          }}
          transition={{
            duration: 4,
            repeat: Number.POSITIVE_INFINITY,
            ease: [0.4, 0, 0.6, 1],
          }}
        />

        {/* Accent particles - Light mode */}
        <motion.div
          className="absolute inset-0 rounded-full dark:hidden"
          style={{
            background: `conic-gradient(from 270deg, transparent 0deg, hsl(var(--primary) / 0.4) 20deg, transparent 40deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 61%, black 62%, black 63%, transparent 64%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 61%, black 62%, black 63%, transparent 64%)`,
            opacity: 0.5,
          }}
          animate={{
            rotate: [0, 360],
          }}
          transition={{
            duration: 3.5,
            repeat: Number.POSITIVE_INFINITY,
            ease: "linear",
          }}
        />

        {/* Dark mode variants */}
        <motion.div
          className="absolute inset-0 rounded-full hidden dark:block"
          style={{
            background: `conic-gradient(from 0deg, transparent 0deg, hsl(var(--primary)) 90deg, transparent 180deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 35%, black 37%, black 39%, transparent 41%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 35%, black 37%, black 39%, transparent 41%)`,
            opacity: 0.8,
          }}
          animate={{
            rotate: [0, 360],
          }}
          transition={{
            duration: 3,
            repeat: Number.POSITIVE_INFINITY,
            ease: "linear",
          }}
        />

        <motion.div
          className="absolute inset-0 rounded-full hidden dark:block"
          style={{
            background: `conic-gradient(from 0deg, transparent 0deg, hsl(var(--primary)) 120deg, hsl(var(--primary) / 0.5) 240deg, transparent 360deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 42%, black 44%, black 48%, transparent 50%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 42%, black 44%, black 48%, transparent 50%)`,
            opacity: 0.9,
          }}
          animate={{
            rotate: [0, 360],
          }}
          transition={{
            duration: 2.5,
            repeat: Number.POSITIVE_INFINITY,
            ease: [0.4, 0, 0.6, 1],
          }}
        />

        <motion.div
          className="absolute inset-0 rounded-full hidden dark:block"
          style={{
            background: `conic-gradient(from 180deg, transparent 0deg, hsl(var(--primary) / 0.6) 45deg, transparent 90deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 52%, black 54%, black 56%, transparent 58%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 52%, black 54%, black 56%, transparent 58%)`,
            opacity: 0.35,
          }}
          animate={{
            rotate: [0, -360],
          }}
          transition={{
            duration: 4,
            repeat: Number.POSITIVE_INFINITY,
            ease: [0.4, 0, 0.6, 1],
          }}
        />

        <motion.div
          className="absolute inset-0 rounded-full hidden dark:block"
          style={{
            background: `conic-gradient(from 270deg, transparent 0deg, hsl(var(--primary) / 0.4) 20deg, transparent 40deg)`,
            mask: `radial-gradient(circle at 50% 50%, transparent 61%, black 62%, black 63%, transparent 64%)`,
            WebkitMask: `radial-gradient(circle at 50% 50%, transparent 61%, black 62%, black 63%, transparent 64%)`,
            opacity: 0.5,
          }}
          animate={{
            rotate: [0, 360],
          }}
          transition={{
            duration: 3.5,
            repeat: Number.POSITIVE_INFINITY,
            ease: "linear",
          }}
        />
      </motion.div>

      {showText && (
        <motion.div
          className={cn("text-center", config.spacing, config.maxWidth)}
          initial={{ opacity: 0, y: 12 }}
          animate={{
            opacity: 1,
            y: 0,
          }}
          transition={{
            delay: 0.4,
            duration: 1,
            ease: [0.4, 0, 0.2, 1],
          }}
        >
          <motion.h2
            className={cn(
              config.titleClass,
              "text-foreground/90 font-medium tracking-[-0.02em] leading-[1.15] antialiased"
            )}
            initial={{ opacity: 0, y: 12 }}
            animate={{
              opacity: 1,
              y: 0,
            }}
            transition={{
              delay: 0.6,
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
          >
            <motion.span
              animate={{
                opacity: [0.9, 0.7, 0.9],
              }}
              transition={{
                duration: 3,
                repeat: Number.POSITIVE_INFINITY,
                ease: [0.4, 0, 0.6, 1],
              }}
            >
              {title}
            </motion.span>
          </motion.h2>

          <motion.p
            className={cn(
              config.subtitleClass,
              "text-muted-foreground font-normal tracking-[-0.01em] leading-[1.45] antialiased"
            )}
            initial={{ opacity: 0, y: 8 }}
            animate={{
              opacity: 1,
              y: 0,
            }}
            transition={{
              delay: 0.8,
              duration: 0.8,
              ease: [0.4, 0, 0.2, 1],
            }}
          >
            <motion.span
              animate={{
                opacity: [0.6, 0.4, 0.6],
              }}
              transition={{
                duration: 4,
                repeat: Number.POSITIVE_INFINITY,
                ease: [0.4, 0, 0.6, 1],
              }}
            >
              {subtitle}
            </motion.span>
          </motion.p>
        </motion.div>
      )}
    </div>
  );
}

interface PageLoaderProps {
  title?: string;
  subtitle?: string;
  size?: "sm" | "md" | "lg";
}

function PageLoader({
  title = "Carregando...",
  subtitle = "Aguarde enquanto preparamos tudo para você",
  size = "md",
}: PageLoaderProps) {
  return (
    <div className="flex min-h-[50vh] items-center justify-center">
      <Loader title={title} subtitle={subtitle} size={size} />
    </div>
  );
}

export { Loader, PageLoader };
