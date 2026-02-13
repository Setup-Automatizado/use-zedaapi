import { z } from "zod";

/**
 * Schema for creating a new proxy provider.
 */
export const CreateProviderSchema = z.object({
	name: z.string().min(1, "Provider name is required").max(100),
	providerType: z.string().min(1, "Provider type is required"),
	enabled: z.boolean().default(true),
	priority: z.number().int().min(1).max(999).default(100),
	apiKey: z.string().min(1, "API key is required"),
	apiEndpoint: z
		.string()
		.url("Must be a valid URL")
		.optional()
		.or(z.literal("")),
	maxProxies: z.number().int().min(0).default(0),
	maxInstancesPerProxy: z.number().int().min(1).max(100).default(1),
	countryCodes: z.array(z.string().length(2)).default([]),
	rateLimitRpm: z.number().int().min(1).max(1000).default(60),
});

/**
 * Schema for updating an existing provider (all fields optional).
 */
export const UpdateProviderSchema = z.object({
	name: z.string().min(1).max(100).optional(),
	enabled: z.boolean().optional(),
	priority: z.number().int().min(1).max(999).optional(),
	apiKey: z.string().min(1).optional(),
	apiEndpoint: z.string().url().optional().or(z.literal("")),
	maxProxies: z.number().int().min(0).optional(),
	maxInstancesPerProxy: z.number().int().min(1).max(100).optional(),
	countryCodes: z.array(z.string().length(2)).optional(),
	rateLimitRpm: z.number().int().min(1).max(1000).optional(),
});

/**
 * Schema for assigning a pool proxy to an instance.
 */
export const AssignPoolProxySchema = z.object({
	providerId: z.string().uuid().optional(),
	countryCodes: z.array(z.string().length(2)).optional(),
	noWebsocket: z.boolean().default(false),
	onlyLogin: z.boolean().default(false),
	noMedia: z.boolean().default(false),
});

/**
 * Schema for assigning an instance to a group.
 */
export const AssignGroupSchema = z.object({
	groupId: z.string().uuid("Invalid group ID"),
});

/**
 * Schema for creating a new proxy group.
 */
export const CreateGroupSchema = z.object({
	name: z.string().min(1, "Group name is required").max(100),
	providerId: z.string().uuid().optional(),
	maxInstances: z.number().int().min(1).max(1000).default(10),
	countryCode: z.string().length(2).optional(),
});

/**
 * Schema for bulk-assigning pool proxies.
 */
export const BulkAssignSchema = z.object({
	instanceIds: z.array(z.string().uuid()).optional(),
	providerId: z.string().uuid().optional(),
	countryCodes: z.array(z.string().length(2)).optional(),
});

// Inferred types
export type CreateProviderInput = z.input<typeof CreateProviderSchema>;
export type CreateProviderFormValues = z.infer<typeof CreateProviderSchema>;
export type UpdateProviderInput = z.input<typeof UpdateProviderSchema>;
export type UpdateProviderFormValues = z.infer<typeof UpdateProviderSchema>;
export type AssignPoolProxyInput = z.input<typeof AssignPoolProxySchema>;
export type AssignGroupInput = z.input<typeof AssignGroupSchema>;
export type CreateGroupInput = z.input<typeof CreateGroupSchema>;
export type CreateGroupFormValues = z.infer<typeof CreateGroupSchema>;
export type BulkAssignInput = z.input<typeof BulkAssignSchema>;
export type BulkAssignFormValues = z.infer<typeof BulkAssignSchema>;
