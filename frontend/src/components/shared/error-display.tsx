/**
 * ErrorDisplay Component
 *
 * Displays error messages with optional retry functionality.
 * Supports different error types and formats backend API errors.
 */

import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface ErrorDisplayProps {
  /** Error object or error message */
  error: unknown;
  /** Optional title override */
  title?: string;
  /** Optional retry callback */
  onRetry?: () => void;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Extract error message from various error types
 */
function getErrorMessage(error: unknown): string {
  // Handle string errors
  if (typeof error === "string") {
    return error;
  }

  // Handle Error objects
  if (error instanceof Error) {
    return error.message;
  }

  // Handle RTK Query errors (FetchBaseQueryError)
  if (error && typeof error === "object") {
    // RTK Query error format
    if ("status" in error && "data" in error) {
      const data = error.data as any;

      // Backend API error format: { success: false, error: { code, message } }
      if (data?.error?.message) {
        return data.error.message;
      }

      // Fallback to error message
      if (data?.message) {
        return data.message;
      }
    }

    // Standard error format
    if ("message" in error && typeof error.message === "string") {
      return error.message;
    }
  }

  // Fallback for unknown error types
  return "An unexpected error occurred. Please try again.";
}

export function ErrorDisplay({
  error,
  title = "Error",
  onRetry,
  className
}: ErrorDisplayProps) {
  const errorMessage = getErrorMessage(error);

  return (
    <Alert variant="destructive" className={cn("", className)}>
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription className="mt-2">
        <p>{errorMessage}</p>
        {onRetry && (
          <Button
            variant="outline"
            size="sm"
            onClick={onRetry}
            className="mt-4"
          >
            Try Again
          </Button>
        )}
      </AlertDescription>
    </Alert>
  );
}

/**
 * InlineErrorDisplay Component
 *
 * Compact error display for form fields or inline messages.
 */
export function InlineErrorDisplay({
  error,
  className
}: {
  error: string;
  className?: string
}) {
  if (!error) return null;

  return (
    <p className={cn("text-sm text-destructive flex items-center gap-1", className)}>
      <AlertCircle className="h-3 w-3" />
      {error}
    </p>
  );
}

/**
 * FullPageErrorDisplay Component
 *
 * Centers error message in the middle of the screen.
 * Use for full-page error states.
 */
export function FullPageErrorDisplay({
  error,
  title = "Something went wrong",
  onRetry
}: ErrorDisplayProps) {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-md">
        <ErrorDisplay error={error} title={title} onRetry={onRetry} />
      </div>
    </div>
  );
}
