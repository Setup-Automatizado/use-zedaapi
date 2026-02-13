"use client";

import { Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

export interface WebhookFieldProps {
	name: string;
	label: string;
	description?: string;
	value: string;
	error?: string;
	onChange: (value: string) => void;
	onClear?: () => void;
	disabled?: boolean;
	required?: boolean;
}

/**
 * WebhookField Component
 *
 * Reusable webhook URL input field with validation, clear functionality.
 * Displays validation status and allows clearing the field.
 *
 * @param name - Field name (for form identification)
 * @param label - Field label
 * @param description - Optional field description
 * @param value - Current field value
 * @param error - Validation error message
 * @param onChange - Change handler
 * @param onClear - Clear button handler
 * @param disabled - Disable field
 * @param required - Mark field as required
 */
export function WebhookField({
	name,
	label,
	description,
	value,
	error,
	onChange,
	onClear,
	disabled = false,
	required = false,
}: WebhookFieldProps) {
	const hasValue = value && value.length > 0;
	const isValid = hasValue && !error;
	const isInvalid = hasValue && error;

	return (
		<div className="space-y-2">
			<div className="flex items-center justify-between">
				<Label
					htmlFor={name}
					className={cn(
						required &&
							'after:content-["*"] after:ml-0.5 after:text-destructive',
					)}
				>
					{label}
				</Label>
				{hasValue && !disabled && onClear && (
					<Button
						type="button"
						variant="ghost"
						size="sm"
						onClick={onClear}
						className="h-6 px-2 text-xs"
					>
						Clear
					</Button>
				)}
			</div>

			{description && (
				<p className="text-sm text-muted-foreground">{description}</p>
			)}

			<div className="relative">
				<Input
					id={name}
					name={name}
					type="url"
					placeholder="https://example.com/webhook"
					value={value}
					onChange={(e) => onChange(e.target.value)}
					disabled={disabled}
					className={cn(
						"pr-10",
						isValid && "border-green-500 focus-visible:ring-green-500",
						isInvalid && "border-destructive focus-visible:ring-destructive",
					)}
					aria-invalid={isInvalid ? "true" : undefined}
					aria-describedby={
						error
							? `${name}-error`
							: description
								? `${name}-description`
								: undefined
					}
				/>

				{/* Validation indicator */}
				<div className="absolute right-3 top-1/2 -translate-y-1/2">
					{isValid && (
						<Check className="h-4 w-4 text-green-500" aria-label="Valid URL" />
					)}
					{isInvalid && (
						<X className="h-4 w-4 text-destructive" aria-label="Invalid URL" />
					)}
				</div>
			</div>

			{error && (
				<p
					id={`${name}-error`}
					className="text-sm text-destructive"
					role="alert"
				>
					{error}
				</p>
			)}
		</div>
	);
}
