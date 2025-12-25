/**
 * EmptyState Component
 *
 * Displays empty state messages with optional action buttons.
 * Used for lists, tables, or sections with no data.
 */

import { LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface EmptyStateProps {
  /** Icon component from lucide-react */
  icon?: LucideIcon;
  /** Title text for the empty state */
  title: string;
  /** Description text explaining the empty state */
  description?: string;
  /** Optional action button */
  action?: {
    label: string;
    onClick: () => void;
  };
  /** Additional CSS classes */
  className?: string;
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  className
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center py-12 px-4 text-center",
        className
      )}
    >
      {Icon && (
        <div className="mb-4 rounded-full bg-muted p-3">
          <Icon className="h-6 w-6 text-muted-foreground" />
        </div>
      )}
      <h3 className="text-lg font-semibold">{title}</h3>
      {description && (
        <p className="mt-2 text-sm text-muted-foreground max-w-md">
          {description}
        </p>
      )}
      {action && (
        <Button onClick={action.onClick} className="mt-6">
          {action.label}
        </Button>
      )}
    </div>
  );
}

/**
 * TableEmptyState Component
 *
 * Specialized empty state for tables with consistent styling.
 */
export function TableEmptyState({
  icon: Icon,
  title,
  description,
  action
}: EmptyStateProps) {
  return (
    <div className="flex items-center justify-center py-8">
      <EmptyState
        icon={Icon}
        title={title}
        description={description}
        action={action}
      />
    </div>
  );
}
