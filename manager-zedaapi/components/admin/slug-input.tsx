"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Lock, LockOpen } from "lucide-react";
import { slugify } from "@/lib/slugify";
import { cn } from "@/lib/utils";

interface SlugInputProps {
	value: string;
	onChange: (slug: string) => void;
	sourceValue: string;
	prefix?: string;
	label?: string;
	className?: string;
}

export function SlugInput({
	value,
	onChange,
	sourceValue,
	prefix = "/blog",
	label = "Slug",
	className,
}: SlugInputProps) {
	const [locked, setLocked] = useState(true);
	const prevSourceRef = useRef(sourceValue);

	useEffect(() => {
		if (locked && sourceValue !== prevSourceRef.current) {
			onChange(slugify(sourceValue));
		}
		prevSourceRef.current = sourceValue;
	}, [locked, sourceValue, onChange]);

	const handleManualChange = useCallback(
		(e: React.ChangeEvent<HTMLInputElement>) => {
			onChange(slugify(e.target.value));
		},
		[onChange],
	);

	const toggleLock = useCallback(() => {
		setLocked((prev) => {
			if (prev) return false;
			onChange(slugify(sourceValue));
			return true;
		});
	}, [sourceValue, onChange]);

	return (
		<div className={cn("space-y-1.5", className)}>
			<Label>{label}</Label>
			<div className="flex items-center gap-2">
				<div className="relative flex-1">
					<Input
						value={value}
						onChange={handleManualChange}
						readOnly={locked}
						className={cn(
							"pr-10 font-mono text-sm",
							locked && "bg-muted/50 text-muted-foreground",
						)}
						placeholder="slug-do-post"
					/>
					<Button
						type="button"
						variant="ghost"
						size="icon-xs"
						className="absolute right-2 top-1/2 -translate-y-1/2"
						onClick={toggleLock}
						aria-label={
							locked ? "Desbloquear slug" : "Bloquear slug"
						}
					>
						{locked ? (
							<Lock className="size-3.5 text-muted-foreground" />
						) : (
							<LockOpen className="size-3.5 text-foreground" />
						)}
					</Button>
				</div>
			</div>
			{value && (
				<p className="text-xs text-muted-foreground">
					<span className="text-muted-foreground/70">{prefix}/</span>
					<span className="font-medium text-foreground">{value}</span>
				</p>
			)}
		</div>
	);
}
