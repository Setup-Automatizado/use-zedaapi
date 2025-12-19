/**
 * Generic API Types
 *
 * Common type definitions for API requests, responses, and utilities.
 * Used across all API endpoints for consistency.
 */

/**
 * Standard action result with success/error handling
 * Generic type parameter T represents the success data payload
 */
export interface ActionResult<T = void> {
  /** Operation success status */
  success: boolean;

  /** Success data payload (present when success is true) */
  data?: T;

  /** Single error message (present when success is false) */
  error?: string;

  /** Validation errors by field name (present when success is false) */
  errors?: Record<string, string[]>;
}

/**
 * Pagination parameters for list requests
 */
export interface PaginationParams {
  /** Page number (1-indexed, default: 1) */
  page?: number;

  /** Items per page (default: 10, max: 100) */
  pageSize?: number;

  /** Search query string */
  query?: string;

  /** Sort field name */
  sortBy?: string;

  /** Sort direction */
  sortOrder?: 'asc' | 'desc';
}

/**
 * Paginated response wrapper
 * Generic type parameter T represents the content item type
 */
export interface PaginatedResponse<T> {
  /** Total number of items across all pages */
  total: number;

  /** Total number of pages */
  totalPage: number;

  /** Items per page */
  pageSize: number;

  /** Current page number (1-indexed) */
  page: number;

  /** Items for current page */
  content: T[];
}

/**
 * API error response structure
 */
export interface ApiError {
  /** HTTP status code */
  status: number;

  /** Error message */
  message: string;

  /** Detailed error information */
  details?: string;

  /** Error code for programmatic handling */
  code?: string;

  /** Request ID for debugging */
  requestId?: string;

  /** Timestamp of error occurrence */
  timestamp?: string;

  /** Field-level validation errors */
  validationErrors?: Record<string, string[]>;
}

/**
 * API success response wrapper
 */
export interface ApiResponse<T = unknown> {
  /** Response data */
  data: T;

  /** Response metadata */
  meta?: ResponseMetadata;
}

/**
 * Response metadata
 */
export interface ResponseMetadata {
  /** Request ID for tracing */
  requestId?: string;

  /** Response timestamp */
  timestamp: string;

  /** Response generation time in milliseconds */
  durationMs?: number;

  /** API version */
  version?: string;
}

/**
 * Sort options
 */
export interface SortOptions {
  /** Field to sort by */
  field: string;

  /** Sort direction */
  direction: 'asc' | 'desc';
}

/**
 * Filter options for list endpoints
 */
export interface FilterOptions {
  /** Search query */
  search?: string;

  /** Status filter */
  status?: string | string[];

  /** Date range filter */
  dateRange?: DateRange;

  /** Custom filters */
  [key: string]: unknown;
}

/**
 * Date range filter
 */
export interface DateRange {
  /** Start date (ISO 8601) */
  from: string;

  /** End date (ISO 8601) */
  to: string;
}

/**
 * Batch operation request
 */
export interface BatchRequest<T> {
  /** Operation type */
  operation: string;

  /** Items to process */
  items: T[];

  /** Stop on first error */
  stopOnError?: boolean;
}

/**
 * Batch operation response
 */
export interface BatchResponse<T> {
  /** Total items processed */
  total: number;

  /** Successful operations */
  succeeded: number;

  /** Failed operations */
  failed: number;

  /** Individual results */
  results: BatchResult<T>[];
}

/**
 * Individual batch result
 */
export interface BatchResult<T> {
  /** Result index */
  index: number;

  /** Operation success status */
  success: boolean;

  /** Result data if successful */
  data?: T;

  /** Error message if failed */
  error?: string;
}

/**
 * Type guard to check if action result is successful
 */
export function isSuccess<T>(result: ActionResult<T>): result is ActionResult<T> & { data: T } {
  return result.success === true && result.data !== undefined;
}

/**
 * Type guard to check if action result is error
 */
export function isError<T>(result: ActionResult<T>): result is ActionResult<T> & { error: string } {
  return result.success === false && (result.error !== undefined || result.errors !== undefined);
}

/**
 * Create success result
 */
export function success<T>(data: T): ActionResult<T> {
  return { success: true, data };
}

/**
 * Create error result with single message
 */
export function error<T = void>(message: string): ActionResult<T> {
  return { success: false, error: message };
}

/**
 * Create error result with validation errors
 */
export function validationError<T = void>(errors: Record<string, string[]>): ActionResult<T> {
  return { success: false, errors };
}

/**
 * Extract data from successful result or throw error
 */
export function unwrap<T>(result: ActionResult<T>): T {
  if (!isSuccess(result)) {
    throw new Error(result.error || 'Operation failed');
  }
  return result.data;
}

/**
 * Extract data from successful result or return default value
 */
export function unwrapOr<T>(result: ActionResult<T>, defaultValue: T): T {
  return isSuccess(result) ? result.data : defaultValue;
}

/**
 * Calculate pagination metadata
 */
export function calculatePagination(
  total: number,
  page: number,
  pageSize: number
): Pick<PaginatedResponse<unknown>, 'total' | 'totalPage' | 'page' | 'pageSize'> {
  return {
    total,
    totalPage: Math.ceil(total / pageSize),
    page,
    pageSize,
  };
}

/**
 * Validate pagination parameters
 */
export function validatePaginationParams(params: PaginationParams): PaginationParams {
  const page = Math.max(1, params.page || 1);
  const pageSize = Math.min(100, Math.max(1, params.pageSize || 10));

  return {
    ...params,
    page,
    pageSize,
  };
}

/**
 * HTTP methods
 */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

/**
 * Request configuration
 */
export interface RequestConfig {
  /** HTTP method */
  method?: HttpMethod;

  /** Request headers */
  headers?: Record<string, string>;

  /** Query parameters */
  params?: Record<string, string | number | boolean>;

  /** Request body */
  body?: unknown;

  /** Request timeout in milliseconds */
  timeout?: number;

  /** Enable retry on failure */
  retry?: boolean;

  /** Maximum retry attempts */
  maxRetries?: number;
}

/**
 * API client configuration
 */
export interface ApiClientConfig {
  /** Base API URL */
  baseUrl: string;

  /** Authentication token */
  token?: string;

  /** Default request timeout */
  timeout?: number;

  /** Default headers */
  headers?: Record<string, string>;

  /** Enable request/response logging */
  logging?: boolean;
}
