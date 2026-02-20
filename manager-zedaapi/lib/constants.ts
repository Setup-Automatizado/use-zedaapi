// Plan limits
export const PLAN_LIMITS = {
	starter: {
		maxInstances: 1,
		maxMessagesPerDay: 1000,
		maxContacts: 500,
		hasWebhooks: true,
		hasApiAccess: true,
		hasPrioritySupport: false,
		hasCustomBranding: false,
	},
	professional: {
		maxInstances: 5,
		maxMessagesPerDay: 10000,
		maxContacts: 5000,
		hasWebhooks: true,
		hasApiAccess: true,
		hasPrioritySupport: true,
		hasCustomBranding: false,
	},
	enterprise: {
		maxInstances: 50,
		maxMessagesPerDay: 100000,
		maxContacts: 50000,
		hasWebhooks: true,
		hasApiAccess: true,
		hasPrioritySupport: true,
		hasCustomBranding: true,
	},
} as const;

export type PlanType = keyof typeof PLAN_LIMITS;

// Subscription statuses
export const SUBSCRIPTION_STATUS = {
	ACTIVE: "active",
	PAST_DUE: "past_due",
	CANCELED: "canceled",
	UNPAID: "unpaid",
	TRIALING: "trialing",
	PAUSED: "paused",
	INCOMPLETE: "incomplete",
	INCOMPLETE_EXPIRED: "incomplete_expired",
} as const;

export type SubscriptionStatus =
	(typeof SUBSCRIPTION_STATUS)[keyof typeof SUBSCRIPTION_STATUS];

// Instance statuses
export const INSTANCE_STATUS = {
	CREATING: "creating",
	CONNECTING: "connecting",
	CONNECTED: "connected",
	DISCONNECTED: "disconnected",
	BANNED: "banned",
	DELETED: "deleted",
	ERROR: "error",
} as const;

export type InstanceStatus =
	(typeof INSTANCE_STATUS)[keyof typeof INSTANCE_STATUS];

// Payment methods
export const PAYMENT_METHOD = {
	STRIPE: "stripe",
	PIX: "pix",
	BOLETO: "boleto",
} as const;

export type PaymentMethod =
	(typeof PAYMENT_METHOD)[keyof typeof PAYMENT_METHOD];

// Invoice / NFSe statuses
export const NFSE_STATUS = {
	PENDING: "pending",
	PROCESSING: "processing",
	ISSUED: "issued",
	CANCELED: "canceled",
	ERROR: "error",
} as const;

export type NFSeStatus = (typeof NFSE_STATUS)[keyof typeof NFSE_STATUS];

// Affiliate statuses
export const AFFILIATE_STATUS = {
	PENDING: "pending",
	ACTIVE: "active",
	SUSPENDED: "suspended",
	REJECTED: "rejected",
} as const;

export type AffiliateStatus =
	(typeof AFFILIATE_STATUS)[keyof typeof AFFILIATE_STATUS];

// Routes
export const ROUTES = {
	HOME: "/",
	LOGIN: "/login",
	REGISTER: "/register",
	FORGOT_PASSWORD: "/forgot-password",
	RESET_PASSWORD: "/reset-password",
	VERIFY_EMAIL: "/verify-email",
	WAITLIST: "/waitlist",

	// Dashboard
	DASHBOARD: "/dashboard",
	INSTANCES: "/instances",
	INSTANCE_DETAIL: (id: string) => `/instances/${id}`,
	BILLING: "/billing",
	SETTINGS: "/settings",
	API_KEYS: "/api-keys",

	// Admin
	ADMIN: "/admin",
	ADMIN_USERS: "/admin/users",
	ADMIN_INSTANCES: "/admin/instances",
	ADMIN_SUBSCRIPTIONS: "/admin/subscriptions",
	ADMIN_AFFILIATES: "/admin/affiliates",
	ADMIN_NFSE: "/admin/nfse",
	ADMIN_WAITLIST: "/admin/waitlist",
} as const;

// Cookie names
export const COOKIES = {
	SESSION: "better-auth.session_token",
	THEME: "theme",
	LOCALE: "locale",
} as const;

// Pagination defaults
export const PAGINATION = {
	DEFAULT_PAGE: 1,
	DEFAULT_PAGE_SIZE: 20,
	MAX_PAGE_SIZE: 100,
} as const;

// Rate limit defaults
export const RATE_LIMITS = {
	API: { max: 100, windowMs: 60_000 },
	AUTH: { max: 10, windowMs: 900_000 },
	WEBHOOK: { max: 1000, windowMs: 60_000 },
} as const;
