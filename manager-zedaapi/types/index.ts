export interface ActionResult<T = void> {
	success: boolean;
	data?: T;
	error?: string;
	code?: string;
	errors?: Record<string, string[]>;
}

export interface PaginatedResult<T> {
	items: T[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
	hasMore: boolean;
}

export interface PaginationParams {
	page?: number;
	pageSize?: number;
	search?: string;
	sortBy?: string;
	sortOrder?: "asc" | "desc";
}

export type ServerActionResponse<T = void> = Promise<ActionResult<T>>;

export interface SelectOption {
	label: string;
	value: string;
	disabled?: boolean;
}

export interface DateRange {
	from: Date;
	to: Date;
}

export interface ApiResponse<T = unknown> {
	data: T;
	message?: string;
	meta?: {
		total?: number;
		page?: number;
		pageSize?: number;
	};
}

export interface ApiError {
	error: string;
	code: string;
	statusCode: number;
	details?: Record<string, string[]>;
}
