"use client";

import { Search, X } from "lucide-react";
import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

export interface InstanceFiltersProps {
	status?: string;
	query?: string;
	onStatusChange: (status: string) => void;
	onQueryChange: (query: string) => void;
	onClearFilters?: () => void;
}

const STATUS_OPTIONS = [
	{ value: "all", label: "All" },
	{ value: "connected", label: "Connected" },
	{ value: "disconnected", label: "Disconnected" },
] as const;

export function InstanceFilters({
	status = "all",
	query = "",
	onStatusChange,
	onQueryChange,
	onClearFilters,
}: InstanceFiltersProps) {
	const hasActiveFilters = status !== "all" || query !== "";

	const handleClearFilters = () => {
		if (onClearFilters) {
			onClearFilters();
		} else {
			onStatusChange("all");
			onQueryChange("");
		}
	};

	return (
		<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
			<div className="relative flex-1 max-w-md">
				<Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground pointer-events-none" />
				<Input
					type="search"
					placeholder="Search instances..."
					value={query}
					onChange={(e) => onQueryChange(e.target.value)}
					className="pl-9 pr-9"
				/>
				{query && (
					<Button
						variant="ghost"
						size="icon-sm"
						onClick={() => onQueryChange("")}
						className="absolute right-1 top-1/2 -translate-y-1/2"
					>
						<X className="h-4 w-4" />
						<span className="sr-only">Clear search</span>
					</Button>
				)}
			</div>

			<div className="flex items-center gap-2">
				<Select value={status} onValueChange={onStatusChange}>
					<SelectTrigger className="w-[180px]">
						<SelectValue placeholder="Status" />
					</SelectTrigger>
					<SelectContent>
						{STATUS_OPTIONS.map((option) => (
							<SelectItem key={option.value} value={option.value}>
								{option.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>

				{hasActiveFilters && (
					<Button variant="ghost" size="sm" onClick={handleClearFilters}>
						<X className="h-4 w-4" />
						Clear filters
					</Button>
				)}
			</div>
		</div>
	);
}
