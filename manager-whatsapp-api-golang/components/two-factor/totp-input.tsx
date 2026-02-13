"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

interface TotpInputProps {
	value: string;
	onChange: (value: string) => void;
	disabled?: boolean;
	autoFocus?: boolean;
	className?: string;
}

/**
 * TOTP Input Component
 *
 * A 6-digit OTP input with individual boxes for each digit.
 * Supports auto-focus, paste, and keyboard navigation.
 */
export function TotpInput({
	value,
	onChange,
	disabled = false,
	autoFocus = true,
	className,
}: TotpInputProps) {
	const inputRefs = React.useRef<(HTMLInputElement | null)[]>([]);
	const digits = value.padEnd(6, "").slice(0, 6).split("");

	const focusInput = (index: number) => {
		if (index >= 0 && index < 6) {
			inputRefs.current[index]?.focus();
		}
	};

	const handleChange = (index: number, newValue: string) => {
		// Only allow digits
		const digit = newValue.replace(/\D/g, "").slice(-1);

		const newDigits = [...digits];
		newDigits[index] = digit;
		const newCode = newDigits.join("");
		onChange(newCode);

		// Auto-focus next input
		if (digit && index < 5) {
			focusInput(index + 1);
		}
	};

	const handleKeyDown = (
		index: number,
		e: React.KeyboardEvent<HTMLInputElement>,
	) => {
		if (e.key === "Backspace") {
			if (!digits[index] && index > 0) {
				// Move to previous input if current is empty
				focusInput(index - 1);
			} else {
				// Clear current input
				const newDigits = [...digits];
				newDigits[index] = "";
				onChange(newDigits.join(""));
			}
		} else if (e.key === "ArrowLeft" && index > 0) {
			e.preventDefault();
			focusInput(index - 1);
		} else if (e.key === "ArrowRight" && index < 5) {
			e.preventDefault();
			focusInput(index + 1);
		}
	};

	const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
		e.preventDefault();
		const pastedData = e.clipboardData.getData("text").replace(/\D/g, "");
		if (pastedData) {
			const newCode = pastedData.slice(0, 6);
			onChange(newCode);
			// Focus the next empty input or the last one
			const focusIndex = Math.min(newCode.length, 5);
			focusInput(focusIndex);
		}
	};

	const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
		e.target.select();
	};

	return (
		<div className={cn("flex gap-2 justify-center", className)}>
			{[0, 1, 2, 3, 4, 5].map((index) => (
				<input
					key={index}
					ref={(el) => {
						inputRefs.current[index] = el;
					}}
					type="text"
					inputMode="numeric"
					maxLength={1}
					value={digits[index] || ""}
					onChange={(e) => handleChange(index, e.target.value)}
					onKeyDown={(e) => handleKeyDown(index, e)}
					onPaste={handlePaste}
					onFocus={handleFocus}
					disabled={disabled}
					autoFocus={autoFocus && index === 0}
					aria-label={`Digit ${index + 1} of 6`}
					className={cn(
						"w-12 h-14 text-center text-xl font-semibold",
						"bg-input/30 border-input rounded-xl border",
						"focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
						"transition-colors outline-none",
						"disabled:pointer-events-none disabled:opacity-50",
					)}
				/>
			))}
		</div>
	);
}

export default TotpInput;
