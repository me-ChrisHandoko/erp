/**
 * Invite User Form Component
 *
 * Form for inviting new team members.
 * Sends email invitation with temporary password.
 *
 * Features:
 * - Email, name, and role input
 * - Form validation (email format)
 * - Role selection (ADMIN, STAFF, VIEWER only - no OWNER)
 * - Rate limiting handling (429 error)
 * - Success/error feedback
 */

"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { inviteUserSchema, type InviteUserFormData } from "@/lib/schemas/user.schema";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { useInviteUserMutation } from "@/store/services/tenantApi";
import { LoadingSpinner } from "@/components/shared/loading-spinner";

interface InviteUserFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function InviteUserForm({ onSuccess, onCancel }: InviteUserFormProps) {
  const [inviteUser, { isLoading }] = useInviteUserMutation();

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<InviteUserFormData>({
    resolver: zodResolver(inviteUserSchema),
    defaultValues: {
      role: "STAFF",
    },
  });

  const onSubmit = async (data: InviteUserFormData) => {
    try {
      await inviteUser(data).unwrap();
      toast.success("Invitation sent successfully", {
        description: `An email has been sent to ${data.email}`,
      });
      onSuccess?.();
    } catch (error: unknown) {
      const errorData = error && typeof error === 'object' && 'data' in error ? error.data : null;
      const errorMessage =
        (errorData && typeof errorData === 'object' && 'error' in errorData &&
         errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error
          ? (errorData.error.message as string)
          : null) || "Failed to send invitation";

      // Handle rate limiting
      const errorStatus = error && typeof error === 'object' && 'status' in error ? error.status : null;
      if (errorStatus === 429) {
        toast.error("Rate limit exceeded", {
          description: "Please wait before sending another invitation (max 5 per minute)",
        });
      } else if (errorMessage.includes("user limit")) {
        toast.error("User limit reached", {
          description: "Upgrade your subscription to add more users",
        });
      } else if (errorMessage.includes("already exists")) {
        toast.error("User already exists", {
          description: "This email is already registered in your organization",
        });
      } else {
        toast.error("Failed to send invitation", {
          description: errorMessage,
        });
      }
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Email */}
      <div className="space-y-2">
        <Label htmlFor="email">
          Email <span className="text-red-500">*</span>
        </Label>
        <Input
          id="email"
          type="email"
          {...register("email")}
          placeholder="user@example.com"
          disabled={isLoading}
        />
        {errors.email && (
          <p className="text-sm text-red-500">{errors.email.message}</p>
        )}
      </div>

      {/* Name */}
      <div className="space-y-2">
        <Label htmlFor="name">
          Full Name <span className="text-red-500">*</span>
        </Label>
        <Input
          id="name"
          {...register("name")}
          placeholder="John Doe"
          disabled={isLoading}
        />
        {errors.name && (
          <p className="text-sm text-red-500">{errors.name.message}</p>
        )}
      </div>

      {/* Role */}
      <div className="space-y-2">
        <Label htmlFor="role">
          Role <span className="text-red-500">*</span>
        </Label>
        <Select
          value={watch("role")}
          onValueChange={(value) => setValue("role", value as "ADMIN" | "STAFF" | "VIEWER")}
          disabled={isLoading}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ADMIN">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Admin</span>
                <span className="text-xs text-muted-foreground">
                  Full access to all features and settings
                </span>
              </div>
            </SelectItem>
            <SelectItem value="STAFF">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Staff</span>
                <span className="text-xs text-muted-foreground">
                  Can manage daily operations and transactions
                </span>
              </div>
            </SelectItem>
            <SelectItem value="VIEWER">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Viewer</span>
                <span className="text-xs text-muted-foreground">
                  Read-only access to view data
                </span>
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
        {errors.role && (
          <p className="text-sm text-red-500">{errors.role.message}</p>
        )}
        <p className="text-xs text-muted-foreground">
          Note: OWNER role cannot be assigned through invitation
        </p>
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
            Cancel
          </Button>
        )}
        <Button type="submit" disabled={isLoading}>
          {isLoading ? (
            <>
              <LoadingSpinner size="sm" className="mr-2" />
              Sending Invitation...
            </>
          ) : (
            "Send Invitation"
          )}
        </Button>
      </div>
    </form>
  );
}
