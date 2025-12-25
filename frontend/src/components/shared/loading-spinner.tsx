/**
 * LoadingSpinner Component
 *
 * Displays a loading indicator with optional text.
 * Can be used standalone or as part of loading states.
 */

import { cn } from "@/lib/utils";

interface LoadingSpinnerProps {
  /** Size of the spinner */
  size?: "sm" | "md" | "lg" | "xl";
  /** Optional loading text to display below spinner */
  text?: string;
  /** Additional CSS classes */
  className?: string;
}

const sizeClasses = {
  sm: "h-4 w-4 border-2",
  md: "h-8 w-8 border-2",
  lg: "h-12 w-12 border-3",
  xl: "h-16 w-16 border-4",
};

export function LoadingSpinner({
  size = "md",
  text,
  className
}: LoadingSpinnerProps) {
  return (
    <div className={cn("flex flex-col items-center justify-center gap-2", className)}>
      <div
        className={cn(
          "animate-spin rounded-full border-primary border-t-transparent",
          sizeClasses[size]
        )}
        role="status"
        aria-label="Loading"
      />
      {text && (
        <p className="text-sm text-muted-foreground">{text}</p>
      )}
    </div>
  );
}

/**
 * FullPageLoadingSpinner Component
 *
 * Centers the spinner in the middle of the screen.
 * Use for full-page loading states.
 */
export function FullPageLoadingSpinner({ text }: { text?: string }) {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <LoadingSpinner size="lg" text={text} />
    </div>
  );
}
