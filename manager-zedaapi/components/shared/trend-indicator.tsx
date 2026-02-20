import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { cn } from "@/lib/utils";

interface TrendIndicatorProps {
	value: number;
	className?: string;
}

export function TrendIndicator({ value, className }: TrendIndicatorProps) {
	const isPositive = value > 0;
	const isNeutral = value === 0;
	const Icon = isNeutral ? Minus : isPositive ? TrendingUp : TrendingDown;

	return (
		<span
			className={cn(
				"inline-flex items-center gap-1 text-xs font-medium",
				isPositive && "text-primary",
				!isPositive && !isNeutral && "text-destructive",
				isNeutral && "text-muted-foreground",
				className,
			)}
		>
			<Icon className="size-3" />
			{isNeutral ? "0%" : `${isPositive ? "+" : ""}${value.toFixed(1)}%`}
		</span>
	);
}
