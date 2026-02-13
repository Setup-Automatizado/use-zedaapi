/**
 * Access Control and Permissions Configuration
 *
 * Defines roles and permissions for the WhatsApp Manager platform.
 * Uses Better Auth's access control system.
 *
 * @module lib/auth/permissions
 */

import { createAccessControl } from "better-auth/plugins/access";
import { adminAc, defaultStatements } from "better-auth/plugins/admin/access";

/**
 * Permission statements defining available actions per resource.
 * Uses 'as const' for proper TypeScript inference.
 */
export const statement = {
	...defaultStatements,
	invitation: ["create", "list", "delete"],
	settings: ["view", "edit"],
} as const;

/**
 * Access control instance for managing permissions.
 */
export const ac = createAccessControl(statement);

/**
 * Admin role with full permissions.
 * Can manage users, invitations, and all settings.
 */
export const adminRole = ac.newRole({
	invitation: ["create", "list", "delete"],
	settings: ["view", "edit"],
	...adminAc.statements,
});

/**
 * Regular user role with limited permissions.
 * Can only view settings, no admin operations.
 */
export const userRole = ac.newRole({
	settings: ["view"],
});

// Export access control for use in auth configuration
export { ac as accessControl };
