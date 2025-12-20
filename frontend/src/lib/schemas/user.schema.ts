/**
 * User Validation Schemas
 *
 * Zod schemas for user management operations:
 * - Invite user (email, name, phone, role)
 * - Update user role
 */

import { z } from "zod";

// Indonesian phone regex: +628xxx or 08xxx (8-11 digits after prefix)
const phoneRegex = /^(\+628|08)\d{8,11}$/;

/**
 * Invite User Schema
 * Used for inviting new team members
 */
export const inviteUserSchema = z.object({
  email: z
    .string()
    .min(1, "Email is required")
    .email("Invalid email format")
    .max(255, "Email must be less than 255 characters"),

  name: z
    .string()
    .min(1, "Name is required")
    .max(255, "Name must be less than 255 characters")
    .regex(/^[a-zA-Z\s.'-]+$/, "Name can only contain letters, spaces, and . ' -"),

  phone: z
    .string()
    .regex(phoneRegex, "Invalid Indonesian phone number (format: +628xxx or 08xxx)")
    .optional()
    .or(z.literal("")),

  role: z.enum(["ADMIN", "STAFF", "VIEWER"], {
    message: "Invalid role. Must be ADMIN, STAFF, or VIEWER",
  }),
});

export type InviteUserFormData = z.infer<typeof inviteUserSchema>;

/**
 * Update User Role Schema
 * Used for changing existing user roles
 */
export const updateUserRoleSchema = z.object({
  role: z.enum(["ADMIN", "STAFF", "VIEWER"], {
    message: "Invalid role. Must be ADMIN, STAFF, or VIEWER",
  }),
});

export type UpdateUserRoleFormData = z.infer<typeof updateUserRoleSchema>;
