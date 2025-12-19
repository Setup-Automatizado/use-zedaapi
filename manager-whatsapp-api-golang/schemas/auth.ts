import { z } from 'zod';

/**
 * Schema for user login validation
 * Validates email format and password minimum length
 */
export const loginSchema = z.object({
  email: z
    .string({ message: 'Email is required' })
    .email('Invalid email')
    .trim()
    .toLowerCase(),

  password: z
    .string({ message: 'Password is required' })
    .min(8, 'Password must be at least 8 characters'),
});

/**
 * Type inference for LoginSchema
 */
export type LoginInput = z.infer<typeof loginSchema>;

/**
 * Schema for user registration validation
 * Includes password confirmation matching
 */
export const registerSchema = z
  .object({
    name: z
      .string({ message: 'Name is required' })
      .min(2, 'Name must be at least 2 characters')
      .max(100, 'Name is too long')
      .trim(),

    email: z
      .string({ message: 'Email is required' })
      .email('Invalid email')
      .trim()
      .toLowerCase(),

    password: z
      .string({ message: 'Password is required' })
      .min(8, 'Password must be at least 8 characters')
      .max(128, 'Password is too long')
      .regex(
        /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
        'Password must contain uppercase, lowercase and number'
      ),

    confirmPassword: z
      .string({ message: 'Password confirmation is required' }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

/**
 * Type inference for RegisterSchema
 */
export type RegisterInput = z.infer<typeof registerSchema>;

/**
 * Schema for password reset request
 * Validates email for password recovery
 */
export const passwordResetRequestSchema = z.object({
  email: z
    .string({ message: 'Email is required' })
    .email('Invalid email')
    .trim()
    .toLowerCase(),
});

/**
 * Type inference for PasswordResetRequestSchema
 */
export type PasswordResetRequestInput = z.infer<typeof passwordResetRequestSchema>;

/**
 * Schema for password reset confirmation
 * Validates new password and token
 */
export const passwordResetSchema = z
  .object({
    token: z
      .string({ message: 'Token is required' })
      .min(1, 'Invalid token'),

    password: z
      .string({ message: 'Password is required' })
      .min(8, 'Password must be at least 8 characters')
      .max(128, 'Password is too long')
      .regex(
        /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
        'Password must contain uppercase, lowercase and number'
      ),

    confirmPassword: z
      .string({ message: 'Password confirmation is required' }),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

/**
 * Type inference for PasswordResetSchema
 */
export type PasswordResetInput = z.infer<typeof passwordResetSchema>;

/**
 * Schema for updating user profile
 */
export const updateUserProfileSchema = z.object({
  name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(100, 'Name is too long')
    .trim()
    .optional(),

  email: z
    .string()
    .email('Invalid email')
    .trim()
    .toLowerCase()
    .optional(),
});

/**
 * Type inference for UpdateUserProfileSchema
 */
export type UpdateUserProfileInput = z.infer<typeof updateUserProfileSchema>;

/**
 * Schema for changing user password
 * Requires current password for security
 */
export const changePasswordSchema = z
  .object({
    currentPassword: z
      .string({ message: 'Current password is required' })
      .min(1, 'Current password is required'),

    newPassword: z
      .string({ message: 'New password is required' })
      .min(8, 'Password must be at least 8 characters')
      .max(128, 'Password is too long')
      .regex(
        /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
        'Password must contain uppercase, lowercase and number'
      ),

    confirmNewPassword: z
      .string({ message: 'Password confirmation is required' }),
  })
  .refine((data) => data.newPassword === data.confirmNewPassword, {
    message: 'Passwords do not match',
    path: ['confirmNewPassword'],
  })
  .refine((data) => data.currentPassword !== data.newPassword, {
    message: 'New password must be different from current password',
    path: ['newPassword'],
  });

/**
 * Type inference for ChangePasswordSchema
 */
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;

/**
 * Schema for email verification token
 */
export const emailVerificationSchema = z.object({
  token: z
    .string({ message: 'Token is required' })
    .min(1, 'Invalid token'),
});

/**
 * Type inference for EmailVerificationSchema
 */
export type EmailVerificationInput = z.infer<typeof emailVerificationSchema>;

/**
 * Schema for two-factor authentication setup
 */
export const twoFactorSetupSchema = z.object({
  code: z
    .string({ message: 'Code is required' })
    .length(6, 'Code must be 6 digits')
    .regex(/^\d{6}$/, 'Code must contain only numbers'),
});

/**
 * Type inference for TwoFactorSetupSchema
 */
export type TwoFactorSetupInput = z.infer<typeof twoFactorSetupSchema>;

/**
 * Schema for two-factor authentication verification
 */
export const twoFactorVerificationSchema = z.object({
  code: z
    .string({ message: 'Code is required' })
    .length(6, 'Code must be 6 digits')
    .regex(/^\d{6}$/, 'Code must contain only numbers'),

  trustDevice: z.boolean().default(false),
});

/**
 * Type inference for TwoFactorVerificationSchema
 */
export type TwoFactorVerificationInput = z.infer<typeof twoFactorVerificationSchema>;
